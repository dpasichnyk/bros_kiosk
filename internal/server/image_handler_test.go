package server

import (
	"encoding/json"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bros_kiosk/internal/cache"
	"bros_kiosk/internal/config"
	"bros_kiosk/internal/renderer"
	"bros_kiosk/pkg/fetcher"
)

func TestImageHandler_Basic(t *testing.T) {
	r, _ := renderer.NewGGRenderer()
	srv := &DashboardServer{
		config: &config.Config{
			UI: config.UIConfig{
				Locale:     "en-US",
				TimeFormat: "24h",
			},
		},
		cache:         cache.New(),
		state:         make(map[string]fetcher.Result),
		imageRenderer: r,
	}

	req := httptest.NewRequest("GET", "/dashboard/image?w=320&h=240&format=png", nil)
	w := httptest.NewRecorder()

	srv.ImageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ImageHandler() status = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %s, want image/png", ct)
	}

	_, err := png.Decode(w.Body)
	if err != nil {
		t.Errorf("Failed to decode PNG: %v", err)
	}
}

func TestImageHandler_JPEG(t *testing.T) {
	r, _ := renderer.NewGGRenderer()
	srv := &DashboardServer{
		config: &config.Config{
			UI: config.UIConfig{},
		},
		cache:         cache.New(),
		state:         make(map[string]fetcher.Result),
		imageRenderer: r,
	}

	req := httptest.NewRequest("GET", "/dashboard/image?w=320&h=240&format=jpg", nil)
	w := httptest.NewRecorder()

	srv.ImageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ImageHandler() status = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %s, want image/jpeg", ct)
	}
}

func TestImageHandler_NoRenderer(t *testing.T) {
	srv := &DashboardServer{
		config: &config.Config{},
		cache:  cache.New(),
		state:  make(map[string]fetcher.Result),
	}

	req := httptest.NewRequest("GET", "/dashboard/image", nil)
	w := httptest.NewRecorder()

	srv.ImageHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ImageHandler() status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestCollectDashboardData(t *testing.T) {
	r, _ := renderer.NewGGRenderer()
	srv := &DashboardServer{
		config: &config.Config{
			UI: config.UIConfig{
				Locale:     "en-US",
				TimeFormat: "12h",
			},
			Sections: []config.Section{
				{ID: "weather", Type: "weather"},
				{ID: "news", Type: "rss"},
				{ID: "my-calendar", Type: "calendar"},
			},
		},
		cache:         cache.New(),
		imageRenderer: r,
		state: map[string]fetcher.Result{
			"weather": {
				FetcherName: "weather",
				Data: &fetcher.WeatherData{
					Temp:        20.5,
					Description: "Sunny",
					Icon:        "01d",
					City:        "Munich",
				},
			},
			"news": {
				FetcherName: "news",
				Data: &fetcher.RSSData{
					Items: []fetcher.RSSItem{
						{Title: "Test News", Summary: "Summary", PubDate: time.Now().Format(time.RFC1123Z)},
					},
				},
			},
			"my-calendar": {
				FetcherName: "my-calendar",
				Data: &fetcher.CalendarData{
					Events: []fetcher.CalendarEvent{
						{Summary: "Event 1", Start: time.Now(), AllDay: false},
					},
				},
			},
		},
	}

	data := srv.collectDashboardData(1920, 1080)

	if data.Locale != "en-US" {
		t.Errorf("Locale = %s, want en-US", data.Locale)
	}
	if data.TimeFormat != "12h" {
		t.Errorf("TimeFormat = %s, want 12h", data.TimeFormat)
	}
	if data.Weather == nil {
		t.Error("Weather is nil, expected data")
	} else if data.Weather.City != "Munich" {
		t.Errorf("Weather.City = %s, want Munich", data.Weather.City)
	}
	if len(data.News) != 1 {
		t.Errorf("len(News) = %d, want 1", len(data.News))
	}
	if len(data.Calendar) != 1 {
		t.Errorf("len(Calendar) = %d, want 1", len(data.Calendar))
	}
}

func TestImageHandler_ResolutionLimits(t *testing.T) {
	r, _ := renderer.NewGGRenderer()
	srv := &DashboardServer{
		config:        &config.Config{},
		cache:         cache.New(),
		state:         make(map[string]fetcher.Result),
		imageRenderer: r,
	}

	req := httptest.NewRequest("GET", "/dashboard/image?w=9999&h=9999", nil)
	w := httptest.NewRecorder()

	srv.ImageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ImageHandler() status = %d, want %d", w.Code, http.StatusOK)
	}

	img, _ := png.Decode(w.Body)
	bounds := img.Bounds()
	if bounds.Dx() > 4096 || bounds.Dy() > 4096 {
		t.Errorf("Image exceeds max size: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestCollectDashboardData_Empty(t *testing.T) {
	srv := &DashboardServer{
		config: &config.Config{
			UI: config.UIConfig{
				Locale: "de-DE",
			},
		},
		cache: cache.New(),
		state: make(map[string]fetcher.Result),
	}

	data := srv.collectDashboardData(1920, 1080)

	if data.Weather != nil {
		t.Error("Weather should be nil for empty state")
	}
	if len(data.News) != 0 {
		t.Error("News should be empty for empty state")
	}
	if len(data.Calendar) != 0 {
		t.Error("Calendar should be empty for empty state")
	}
}

func TestImageHandler_JSON_Not_Supported(t *testing.T) {
	r, _ := renderer.NewGGRenderer()
	srv := &DashboardServer{
		config:        &config.Config{},
		cache:         cache.New(),
		state:         make(map[string]fetcher.Result),
		imageRenderer: r,
	}

	req := httptest.NewRequest("GET", "/dashboard/image?format=json", nil)
	w := httptest.NewRecorder()

	srv.ImageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ImageHandler() status = %d, want %d (defaults to PNG)", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %s, want image/png (unknown format defaults to PNG)", ct)
	}

	_ = json.RawMessage{}
}
