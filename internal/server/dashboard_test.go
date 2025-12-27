package server

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bros_kiosk/internal/config"
)

func TestDashboardHandler(t *testing.T) {
	cfg := &config.Config{
		Sections: []config.Section{
			{ID: "weather", Type: "widget", Style: "default"},
		},
	}

	srv := New(cfg)

	req, err := http.NewRequest("GET", "/dashboard", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "<title>Bros Kiosk</title>") {
		t.Error("body does not contain expected title")
	}

	if !strings.Contains(rr.Body.String(), "id=\"weather\"") {
		t.Error("body does not contain expected section ID")
	}
}

func TestDashboardHandler_Error(t *testing.T) {
	srv := &DashboardServer{
		templates: template.Must(template.New("invalid").Parse("{{.NonExistentField}}")),
		config:    &config.Config{},
	}

	req, err := http.NewRequest("GET", "/dashboard", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.DashboardHandler(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}
