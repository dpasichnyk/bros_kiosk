package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WeatherData represents the normalized weather information for the dashboard.
type WeatherData struct {
        Temp          float64 `json:"temp"`
        Humidity      int     `json:"humidity"`
        Description   string  `json:"description"`
        Icon          string  `json:"icon"`
        City          string  `json:"city"`
        SetupRequired bool    `json:"setup_required"`
}
// WeatherFetcher implements the Fetcher interface for OpenWeatherMap.
type WeatherFetcher struct {
	name    string
	apiKey  string
	city    string
	units   string
	baseURL string
	client  *http.Client
}

// NewWeatherFetcher creates a new instance of WeatherFetcher.
func NewWeatherFetcher(apiKey, city, units, baseURL string) *WeatherFetcher {
	if baseURL == "" {
		baseURL = "https://api.openweathermap.org/data/2.5/weather"
	}
	return &WeatherFetcher{
		name:    "weather",
		apiKey:  apiKey,
		city:    city,
		units:   units,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// SetName overrides the default name.
func (f *WeatherFetcher) SetName(name string) {
	f.name = name
}

// Name returns the fetcher name.
func (f *WeatherFetcher) Name() string {
	return f.name
}

// Fetch retrieves weather data from the API.
func (f *WeatherFetcher) Fetch(ctx context.Context) (interface{}, error) {
        if f.apiKey == "" {
                return &WeatherData{
                        Description:   "Setup Required",
                        SetupRequired: true,
                }, nil
        }

        url := fmt.Sprintf("%s?q=%s&appid=%s&units=%s", f.baseURL, f.city, f.apiKey, f.units)

        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
                return nil, fmt.Errorf("failed to create request: %w", err)
        }

        resp, err := f.client.Do(req)
        if err != nil {
                return nil, fmt.Errorf("request failed: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusUnauthorized {
                return &WeatherData{
                        Description:   "Setup Required",
                        SetupRequired: true,
                }, nil
        }

        if resp.StatusCode != http.StatusOK {
                return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
        }
	var owm struct {
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&owm); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(owm.Weather) == 0 {
		return nil, fmt.Errorf("no weather data in response")
	}

	return &WeatherData{
		Temp:        owm.Main.Temp,
		Humidity:    owm.Main.Humidity,
		Description: owm.Weather[0].Description,
		Icon:        owm.Weather[0].Icon,
		City:        owm.Name,
	}, nil
}
