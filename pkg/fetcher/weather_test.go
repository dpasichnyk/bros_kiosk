package fetcher

import (
        "context"
        "net/http"
        "net/http/httptest"
        "strings"
        "testing"
        "time"
)

func TestWeatherFetcher(t *testing.T) {
        // Mock OpenWeatherMap response
        mockResponse := `{
                "main": {"temp": 20.5, "humidity": 50},
                "weather": [{"description": "clear sky", "icon": "01d"}],
                "name": "London"
        }`

        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusOK)
                w.Write([]byte(mockResponse))
        }))
        defer server.Close()

        wf := NewWeatherFetcher("fake-key", "London", "metric", server.URL)

        if wf.Name() != "weather" {
                t.Errorf("Expected name 'weather', got '%s'", wf.Name())
        }

        wf.SetName("custom-weather")
        if wf.Name() != "custom-weather" {
                t.Errorf("Expected name 'custom-weather', got '%s'", wf.Name())
        }

        data, err := wf.Fetch(context.Background())
        if err != nil {
                t.Fatalf("Fetch failed: %v", err)
        }

        weather, ok := data.(*WeatherData)
        if !ok {
                t.Fatalf("Expected *WeatherData, got %T", data)
        }

        if weather.Temp != 20.5 {
                t.Errorf("Expected temp 20.5, got %v", weather.Temp)
        }
        if weather.Description != "clear sky" {
                t.Errorf("Expected description 'clear sky', got '%s'", weather.Description)
        }
}

func TestWeatherFetcher_SetupRequired(t *testing.T) {
        // Case 1: Empty API key
        wf := NewWeatherFetcher("", "London", "metric", "")

        data, err := wf.Fetch(context.Background())
        if err != nil {
                t.Fatalf("Fetch should not return error for empty API key, but got: %v", err)
        }

        weather, ok := data.(*WeatherData)
        if !ok {
                t.Fatalf("Expected *WeatherData, got %T", data)
        }

        if !weather.SetupRequired {
                t.Error("Expected SetupRequired to be true")
        }
        if weather.Description != "Setup Required" {
                t.Errorf("Expected description 'Setup Required', got '%s'", weather.Description)
        }

        // Case 2: Invalid API key (401)
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"message": "invalid api key"}`))
        }))
        defer server.Close()

        wf2 := NewWeatherFetcher("invalid-key", "London", "metric", server.URL)
        data2, err2 := wf2.Fetch(context.Background())
        if err2 != nil {
                t.Fatalf("Fetch should not return error for 401, but got: %v", err2)
        }

        weather2, ok := data2.(*WeatherData)
        if !ok {
                t.Fatalf("Expected *WeatherData, got %T", data2)
        }

        if !weather2.SetupRequired {
                t.Error("Expected SetupRequired to be true for 401")
        }
}

func TestWeatherFetcher_Errors(t *testing.T) {
        tests := []struct {
                name           string
                status         int
                response       string
                expectedErrMsg string
        }{
                {
                        name:           "Internal Server Error",
                        status:         http.StatusInternalServerError,
                        response:       `error`,
                        expectedErrMsg: "unexpected status code: 500",
                },
                {
                        name:           "Invalid JSON",
                        status:         http.StatusOK,
                        response:       `{invalid`,
                        expectedErrMsg: "failed to decode response",
                },
                {
                        name:           "Empty Weather Array",
                        status:         http.StatusOK,
                        response:       `{"main": {"temp": 10}, "weather": [], "name": "London"}`,
                        expectedErrMsg: "no weather data in response",
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                                w.WriteHeader(tt.status)
                                w.Write([]byte(tt.response))
                        }))
                        defer server.Close()

                        wf := NewWeatherFetcher("key", "City", "metric", server.URL)
                        _, err := wf.Fetch(context.Background())

                        if err == nil {
                                t.Fatal("Expected error, got nil")
                        }
                        if !strings.Contains(err.Error(), tt.expectedErrMsg) {
                                t.Errorf("Expected error containing '%s', got: %v", tt.expectedErrMsg, err)
                        }
                })
        }
}

func TestWeatherManagerIntegration(t *testing.T) {
        mockResponse := `{"main": {"temp": 15}, "weather": [{"description": "cloudy"}], "name": "Berlin"}`
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
                w.Write([]byte(mockResponse))
        }))
        defer server.Close()

        manager := NewManager()
        wf := NewWeatherFetcher("key", "Berlin", "metric", server.URL)
        manager.Register(wf, 100*time.Millisecond)

        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        go manager.Start(ctx)

        select {
        case res := <-manager.Updates():
                if res.FetcherName != "weather" {
                        t.Errorf("Expected fetcher name 'weather', got '%s'", res.FetcherName)
                }
                weather := res.Data.(*WeatherData)
                if weather.City != "Berlin" {
                        t.Errorf("Expected city 'Berlin', got '%s'", weather.City)
                }
        case <-time.After(500 * time.Millisecond):
                t.Fatal("Timed out waiting for weather update via manager")
        }
}