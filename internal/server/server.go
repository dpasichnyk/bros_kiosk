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

	manager    *fetcher.Manager
	state      map[string]fetcher.Result
	mu         sync.RWMutex
	imageCache *images.DiskCache
	scannerMgr *scanner.Manager
}

func New(cfg *config.Config) *DashboardServer {
	mux := http.NewServeMux()

	// Parse templates from embedded FS
	tmpl := template.Must(template.ParseFS(assets.FS, "templates/*.html"))

	// Initialize Image Cache
	// Use local directory for persistent cache to avoid re-processing on restart
	imgCache, err := images.NewDiskCache("./kiosk_cache")
	if err != nil {
		panic(err)
	}

	// Initialize Scanners
	var scanners []scanner.Scanner
	for _, src := range cfg.Slideshow.Sources {
		if src.Type == "local" {
			scanners = append(scanners, scanner.NewLocalScanner(src.Path))
		} else if src.Type == "s3" {
			cfg, err := awsconfig.LoadDefaultConfig(context.Background())
			if err != nil {
				slog.Error("Unable to load SDK config, s3 scanner disabled", "error", err)
				continue
			}
			client := s3.NewFromConfig(cfg)
			scanners = append(scanners, scanner.NewS3Scanner(client, src.Bucket, src.Prefix))
		}
	}
	scanMgr := scanner.NewManager(scanners...)

	srv := &DashboardServer{
		config:     cfg,
		stopCh:     make(chan os.Signal, 1),
		templates:  tmpl,
		cache:      cache.New(),
		manager:    fetcher.NewManager(),
		state:      make(map[string]fetcher.Result),
		imageCache: imgCache,
		scannerMgr: scanMgr,
	}

	// Start initial scan in background
	go func() {
		if err := scanMgr.Scan(context.Background()); err != nil {
			slog.Error("Initial photo scan failed", "error", err)
		}
	}()

	// Register fetchers from config
	for _, sec := range cfg.Sections {
		// Determine interval
		interval := 15 * time.Minute // Default
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
				wf.SetName("weather")
				srv.manager.RegisterWithBackoff(wf, interval, 5*time.Second, 1*time.Hour)
			}
		case "rss":
			if sec.RSS != nil {
				rf := fetcher.NewRSSFetcher("news", sec.RSS.URL)
				rf.SetName("news")
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
	mux.HandleFunc("/api/updates", srv.UpdateHandler)
	mux.HandleFunc("/api/photos", srv.PhotosListHandler)
	mux.HandleFunc("/assets/photos/", srv.AssetHandler)

	// Serve static files from the "static" subdirectory of embedded FS
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

	for _, section := range s.config.Sections {
		switch section.Region {
		case "top-left":
			layout.TopLeft = append(layout.TopLeft, section)
		case "top-right":
			layout.TopRight = append(layout.TopRight, section)
		case "center":
			layout.Center = append(layout.Center, section)
		case "bottom-left":
			layout.BottomLeft = append(layout.BottomLeft, section)
		case "bottom-right":
			layout.BottomRight = append(layout.BottomRight, section)
		default:
			// Default to center if unknown? Or maybe log warning.
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
	// Listen for OS signals
	signal.Notify(s.stopCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch listener for fetcher updates
	go s.listenForUpdates(ctx)

	// Start FetcherManager
	go s.manager.Start(ctx)

	// Run server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	// Wait for signal
	<-s.stopCh

	// Create shutdown context with timeout
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
