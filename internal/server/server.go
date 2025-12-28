package server

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"bros_kiosk/assets"
	"bros_kiosk/internal/cache"
	"bros_kiosk/internal/config"
	"bros_kiosk/internal/images"
	"bros_kiosk/internal/renderer"
	"bros_kiosk/internal/scanner"
	"bros_kiosk/pkg/fetcher"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func init() {
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
}

type DashboardServer struct {
	server    *http.Server
	config    *config.Config
	stopCh    chan os.Signal
	templates *template.Template
	cache     *cache.Cache

	manager       *fetcher.Manager
	state         map[string]fetcher.Result
	mu            sync.RWMutex
	imageCache    *images.DiskCache
	scannerMgr    *scanner.Manager
	imageRenderer renderer.Renderer
}

func New(cfg *config.Config) *DashboardServer {
	mux := http.NewServeMux()

	tmpl := template.Must(template.ParseFS(assets.FS, "templates/*.html"))

	imgCache, err := images.NewDiskCache("./kiosk_cache")
	if err != nil {
		panic(err)
	}

	var scanners []scanner.Scanner
	for _, src := range cfg.Slideshow.Sources {
		if src.Type == "local" {
			scanners = append(scanners, scanner.NewLocalScanner(src.Path))
		} else if src.Type == "s3" {
			awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
			if err != nil {
				slog.Error("Unable to load SDK config, s3 scanner disabled", "error", err)
				continue
			}
			client := s3.NewFromConfig(awsCfg)
			scanners = append(scanners, scanner.NewS3Scanner(client, src.Bucket, src.Prefix))
		}
	}
	scanMgr := scanner.NewManager(scanners...)

	ggRenderer, err := renderer.NewGGRenderer()
	if err != nil {
		slog.Error("Failed to initialize image renderer", "error", err)
	}

	var imageRenderer renderer.Renderer = ggRenderer
	if ggRenderer != nil {
		cacheTTL := 5 * time.Second
		if cfg.Server.UpdateInterval != "" {
			if d, err := time.ParseDuration(cfg.Server.UpdateInterval); err == nil {
				cacheTTL = d
			}
		}
		imageRenderer = renderer.NewCachedRenderer(ggRenderer, cacheTTL)
	}

	srv := &DashboardServer{
		config:        cfg,
		stopCh:        make(chan os.Signal, 1),
		templates:     tmpl,
		cache:         cache.New(),
		manager:       fetcher.NewManager(),
		state:         make(map[string]fetcher.Result),
		imageCache:    imgCache,
		scannerMgr:    scanMgr,
		imageRenderer: imageRenderer,
	}

	go func() {
		if err := scanMgr.Scan(context.Background()); err != nil {
			slog.Error("Initial photo scan failed", "error", err)
		}
	}()

	for _, sec := range cfg.Sections {
		interval := 15 * time.Minute
		if sec.Type == "weather" {
			interval = 10 * time.Minute
		}

		if sec.Interval != "" {
			if d, err := time.ParseDuration(sec.Interval); err == nil {
				interval = d
			}
		}

		switch sec.Type {
		case "weather":
			if sec.Weather != nil {
				wf := fetcher.NewWeatherFetcher(sec.Weather.APIKey, sec.Weather.City, sec.Weather.Units, sec.Weather.BaseURL)
				wf.SetName(sec.ID)
				srv.manager.RegisterWithBackoff(wf, interval, 5*time.Second, 1*time.Hour)
			}
		case "rss":
			if sec.RSS != nil {
				rf := fetcher.NewRSSFetcher(sec.ID, sec.RSS.URL)
				if namer, ok := interface{}(rf).(interface{ SetName(string) }); ok {
					namer.SetName(sec.ID)
				}
				srv.manager.RegisterWithBackoff(rf, interval, 5*time.Second, 1*time.Hour)
			}
		case "calendar":
			if len(sec.Calendars) > 0 {
				fetchers := make([]fetcher.Fetcher, 0, len(sec.Calendars))
				for _, cal := range sec.Calendars {
					var f fetcher.Fetcher
					switch cal.Type {
					case "ical":
						f = fetcher.NewICalFetcher(cal.Name, cal.URL)
					case "caldav":
						f = fetcher.NewCalDAVFetcher(cal.Name, cal.URL, cal.Username, cal.Password)
					}
					if f != nil {
						fetchers = append(fetchers, f)
					}
				}

				if len(fetchers) > 0 {
					aggregator := fetcher.NewCalendarAggregator(sec.ID, fetchers)
					srv.manager.RegisterWithBackoff(aggregator, interval, 10*time.Second, 1*time.Hour)
				}
			}
		}
	}

	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/dashboard", srv.DashboardHandler)
	mux.HandleFunc("/dashboard/image", srv.ImageHandler)
	mux.HandleFunc("/api/updates", srv.UpdateHandler)
	mux.HandleFunc("/api/photos", srv.PhotosListHandler)
	mux.HandleFunc("/assets/photos/", srv.AssetHandler)

	staticFS, err := fs.Sub(assets.FS, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	srv.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: mux,
	}

	return srv
}

func (s *DashboardServer) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	layout := struct {
		TopLeft     []config.Section
		TopRight    []config.Section
		Center      []config.Section
		BottomLeft  []config.Section
		BottomRight []config.Section
	}{}

	regionMap := map[string]*[]config.Section{
		"top-left":     &layout.TopLeft,
		"top-right":    &layout.TopRight,
		"center":       &layout.Center,
		"bottom-left":  &layout.BottomLeft,
		"bottom-right": &layout.BottomRight,
	}

	for _, section := range s.config.Sections {
		if ptr, ok := regionMap[section.Region]; ok {
			*ptr = append(*ptr, section)
		} else {
			layout.Center = append(layout.Center, section)
		}
	}

	data := struct {
		Config    *config.Config
		Slideshow config.SlideshowConfig
		Layout    interface{}
	}{
		Config:    s.config,
		Slideshow: s.config.Slideshow,
		Layout:    layout,
	}
	err := s.templates.ExecuteTemplate(w, "dashboard.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *DashboardServer) Start() error {
	signal.Notify(s.stopCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.listenForUpdates(ctx)

	go s.manager.Start(ctx)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	<-s.stopCh

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	return s.server.Shutdown(shutdownCtx)
}

func (s *DashboardServer) listenForUpdates(ctx context.Context) {
	updates := s.manager.Updates()
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-updates:
			s.mu.Lock()
			s.state[result.FetcherName] = result
			s.mu.Unlock()
		}
	}
}

func (s *DashboardServer) Notify(sig os.Signal) {
	s.stopCh <- sig
}
