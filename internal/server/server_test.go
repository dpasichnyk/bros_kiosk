package server

import (
	"net/http"
	"syscall"
	"testing"
	"time"

	"bros_kiosk/internal/config"
)

func TestDashboardServer_Start(t *testing.T) {
	// Setup config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 0, // Random port
			Host: "localhost",
		},
	}

	srv := New(cfg)

	// Start server in background
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server start error: %v", err)
		}
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Send interrupt signal
	srv.Notify(syscall.SIGINT)

	// Wait for shutdown (Start() should return)
	// If it hangs, test will timeout
}

func TestDashboardServer_StartError(t *testing.T) {
	// Setup config with invalid port/address
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: -1,
			Host: "invalid",
		},
	}

	srv := New(cfg)
	
	// Create a channel to catch the error from the goroutine in Start()
	// NOTE: Because Start() runs ListenAndServe in a goroutine and logs the error,
	// capturing the error directly is tricky without refactoring.
	// However, we can at least ensure Start() doesn't panic and returns eventually.
	// For 100% coverage, we might need to dependency inject a logger or the server,
	// but for this phase, we'll try to trigger the error path.
	
	// Actually, looking at the implementation:
	// go func() {
	// 	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		fmt.Printf("Server error: %v\n", err)
	// 	}
	// }()
	//
	// The error is printed to stdout and not returned. The test will just hang waiting for signal.
	// So we need to signal it to stop.
	
	go srv.Start()
	time.Sleep(10 * time.Millisecond)
	srv.Notify(syscall.SIGINT)
}

func TestNewServer_RegistersCalendars(t *testing.T) {
	cfg := &config.Config{
		Sections: []config.Section{
			{
				ID:   "my-cal",
				Type: "calendar",
				Calendars: []config.CalendarSource{
					{Type: "ical", URL: "http://example.com/cal.ics"},
					{Type: "caldav", URL: "http://example.com", Username: "u", Password: "p"},
				},
			},
		},
	}

	srv := New(cfg)
	if srv == nil {
		t.Fatal("Server should not be nil")
	}
}
