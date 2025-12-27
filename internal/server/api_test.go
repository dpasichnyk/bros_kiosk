package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bros_kiosk/internal/config"
	"bros_kiosk/pkg/fetcher"
)

func TestUpdateHandler(t *testing.T) {
	cfg := &config.Config{}
	srv := New(cfg)

	// Populate some state
	srv.mu.Lock()
	srv.state["weather"] = fetcher.Result{
		FetcherName: "weather",
		Data:        map[string]interface{}{"temp": 20},
		Status:      fetcher.Status{IsHealthy: true, LastFetch: time.Now()},
	}
	srv.mu.Unlock()

	req, err := http.NewRequest("GET", "/api/updates", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status ok, got %v", response["status"])
	}
}

func TestUpdateHandler_CachedAndDelta(t *testing.T) {
	cfg := &config.Config{
		Sections: []config.Section{
			{ID: "weather", Type: "widget"},
			{ID: "news", Type: "widget"},
		},
	}
	srv := New(cfg)

	// 1. Initial request
	srv.mu.Lock()
	srv.state["weather"] = fetcher.Result{
		FetcherName: "weather",
		Data:        "initial",
		Status:      fetcher.Status{IsHealthy: true, LastFetch: time.Now()},
	}
	srv.mu.Unlock()

	req1, _ := http.NewRequest("GET", "/api/updates", nil)
	rr1 := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rr1, req1)

	var res1 map[string]interface{}
	json.Unmarshal(rr1.Body.Bytes(), &res1)
	fullHash := res1["hash"].(string)

	// 2. Second request with same hash (304 Not Modified)
	req2, _ := http.NewRequest("GET", "/api/updates", nil)
	req2.Header.Set("X-Dashboard-Hash", fullHash)
	rr2 := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNotModified {
		t.Errorf("expected 304 Not Modified, got %v", rr2.Code)
	}

	// 3. Partial update (simulate change)
	srv.mu.Lock()
	srv.state["weather"] = fetcher.Result{
		FetcherName: "weather",
		Data:        "updated",
		Status:      fetcher.Status{IsHealthy: true, LastFetch: time.Now()},
	}
	srv.mu.Unlock()

	req3, _ := http.NewRequest("GET", "/api/updates", nil)
	req3.Header.Set("X-Dashboard-Hash", fullHash)
	rr3 := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", rr3.Code)
	}

	var res3 map[string]interface{}
	json.Unmarshal(rr3.Body.Bytes(), &res3)

	if res3["hash"] == fullHash {
		t.Error("expected new hash in response")
	}

	updates := res3["updates"].(map[string]interface{})
	if _, ok := updates["weather"]; !ok {
		t.Error("expected weather update")
	}
}

func TestUpdateHandler_HashError(t *testing.T) {
	cfg := &config.Config{
		Sections: []config.Section{
			{ID: "weather"},
		},
	}
	srv := New(cfg)

	// Inject unmarshalable data into state to trigger hashing error
	srv.mu.Lock()
	srv.state["weather"] = fetcher.Result{
		Data: func() {},
	}
	srv.mu.Unlock()

	req, _ := http.NewRequest("GET", "/api/updates", nil)
	rr := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %v", rr.Code)
	}
}
