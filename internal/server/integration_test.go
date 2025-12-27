package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bros_kiosk/internal/config"
	"bros_kiosk/pkg/fetcher"
)

func TestEndToEndIntegration(t *testing.T) {
	// 1. Mock Weather Server
	weatherServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"main": {"temp": 25}, "weather": [{"description": "sunny"}], "name": "Madrid"}`))
	}))
	defer weatherServer.Close()

	// 2. Mock RSS Server
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8" ?><rss version="2.0"><channel><item><title>Test News</title></item></channel></rss>`))
	}))
	defer rssServer.Close()

	// 3. Mock iCal Server
	icalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
SUMMARY:Test Event
DTSTART:20251225T100000Z
END:VEVENT
END:VCALENDAR`))
	}))
	defer icalServer.Close()

	// 4. Create Config
	cfg := &config.Config{
		Server: config.ServerConfig{Host: "localhost", Port: 0},
		Sections: []config.Section{
			{
				ID:   "my-weather",
				Type: "weather",
				Weather: &config.WeatherConfig{
					APIKey:  "key",
					City:    "Madrid",
					Units:   "metric",
					BaseURL: weatherServer.URL,
				},
			},
			{
				ID:   "my-news",
				Type: "rss",
				RSS: &config.RSSConfig{
					URL: rssServer.URL,
				},
			},
			{
				ID:   "my-cal",
				Type: "calendar",
				Calendars: []config.CalendarSource{
					{Type: "ical", URL: icalServer.URL},
				},
			},
		},
	}

	// 5. Start Server components
	srv := New(cfg)
	
	// Speed up intervals for testing by re-registering
	srv.manager = fetcher.NewManager()
	srv.manager.RegisterWithBackoff(fetcher.NewWeatherFetcher("key", "Madrid", "metric", weatherServer.URL), 50*time.Millisecond, 10*time.Millisecond, 50*time.Millisecond)
	srv.manager.RegisterWithBackoff(fetcher.NewRSSFetcher("my-news", rssServer.URL), 50*time.Millisecond, 10*time.Millisecond, 50*time.Millisecond)
	// Register iCal directly as "my-cal" to simulate aggregator result for single source
	// (Aggregator implementation would just wrap this, but for E2E we verify the data flow from state -> API)
	// We'll use ICalFetcher directly to ensure it works.
	icalFetcher := fetcher.NewICalFetcher("my-cal", icalServer.URL)
	srv.manager.RegisterWithBackoff(icalFetcher, 50*time.Millisecond, 10*time.Millisecond, 50*time.Millisecond)

	managerCtx, managerCancel := context.WithCancel(context.Background())
	defer managerCancel()
	
	go srv.listenForUpdates(managerCtx)
	go srv.manager.Start(managerCtx)

	// 6. Poll /api/updates
	timeout := time.After(5 * time.Second)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timed out waiting for integrated updates")
		case <-tick.C:
			req, _ := http.NewRequest("GET", "/api/updates", nil)
			rr := httptest.NewRecorder()
			srv.UpdateHandler(rr, req)

			if rr.Code == http.StatusOK {
				var resp struct {
					Updates map[string]fetcher.Result `json:"updates"`
				}
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					continue
				}

				resWeather, hasWeather := resp.Updates["my-weather"]
				resRSS, hasRSS := resp.Updates["my-news"]
				resCal, hasCal := resp.Updates["my-cal"]

				if hasWeather && hasRSS && hasCal && resWeather.Data != nil && resRSS.Data != nil && resCal.Data != nil {
					// Verify weather data
					wData := resWeather.Data.(map[string]interface{})
					if wData["city"] != "Madrid" {
						t.Errorf("Expected Madrid, got %v", wData["city"])
					}
					// Verify RSS data
					rData := resRSS.Data.(map[string]interface{})
					items := rData["items"].([]interface{})
					if len(items) == 0 {
						t.Errorf("Expected RSS items, got 0")
					}
					// Verify Calendar data
					cData := resCal.Data.(map[string]interface{})
					events := cData["events"].([]interface{})
					if len(events) == 0 {
						t.Errorf("Expected Calendar events, got 0")
					}
					return // Success!
				}
			}
		}
	}
}

func TestPhotoIntegration(t *testing.T) {
	// 1. Setup Temp Photos
	tmpDir, err := os.MkdirTemp("", "e2e_photos")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	photoPath := filepath.Join(tmpDir, "test.jpg")
	createTestImage(t, photoPath)

	// 2. Create Config
	cfg := &config.Config{
		Slideshow: config.SlideshowConfig{
			Sources: []config.SourceConfig{
				{Type: "local", Path: tmpDir},
			},
			TargetResolution: config.Resolution{Width: 50, Height: 50},
		},
	}

	// 3. Start Server
	srv := New(cfg)

	// 4. Force Scan
	if err := srv.scannerMgr.Scan(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 5. Check /api/photos
	reqList := httptest.NewRequest("GET", "/api/photos", nil)
	rrList := httptest.NewRecorder()
	srv.PhotosListHandler(rrList, reqList)

	if rrList.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rrList.Code)
	}

	var listResp struct {
		Photos []string `json:"photos"`
	}
	json.Unmarshal(rrList.Body.Bytes(), &listResp)

	if len(listResp.Photos) == 0 {
		t.Fatal("Expected 1 photo, got 0")
	}

	// 6. Check /assets/photos/
	encodedPath := url.QueryEscape(listResp.Photos[0])
	reqAsset := httptest.NewRequest("GET", "/assets/photos/"+encodedPath, nil)
	rrAsset := httptest.NewRecorder()
	srv.AssetHandler(rrAsset, reqAsset)

	if rrAsset.Code != http.StatusOK {
		t.Fatalf("Expected 200 for asset, got %d", rrAsset.Code)
	}

	if rrAsset.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("Expected image/jpeg, got %s", rrAsset.Header().Get("Content-Type"))
	}
}
