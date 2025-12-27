package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	yamlData := `
server:
  port: 8080
  host: "localhost"
  update_interval: "5s"
sections:
  - id: "weather"
    type: "widget"
    style: "default"
`
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", cfg.Server.Host)
	}
	if cfg.Server.UpdateInterval != "5s" {
		t.Errorf("Expected interval 5s, got %s", cfg.Server.UpdateInterval)
	}
	if len(cfg.Sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(cfg.Sections))
	}
}

func TestEnvVarSubstitution(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	yamlData := `
server:
  port: ${SERVER_PORT}
`
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	t.Run("FileNotFound", func(t *testing.T) {
		_, err := Load("nonexistent.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent file, got nil")
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "invalid*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte("invalid: yaml: :")); err != nil {
			t.Fatal(err)
		}
		tmpFile.Close()

		_, err = Load(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})
}

func TestLoadCalendarConfig(t *testing.T) {
	yamlData := `
sections:
  - id: "my-calendar"
    type: "calendar"
    calendars:
      - type: "caldav"
        url: "https://caldav.example.com"
        username: "user"
        password: "password"
        color: "#ff0000"
      - type: "ical"
        url: "https://example.com/calendar.ics"
        name: "Holidays"
`
	tmpFile, err := os.CreateTemp("", "calendar_config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(cfg.Sections))
	}

	sec := cfg.Sections[0]
	if len(sec.Calendars) != 2 {
		t.Fatalf("Expected 2 calendar sources, got %d", len(sec.Calendars))
	}

	cal1 := sec.Calendars[0]
	if cal1.Type != "caldav" || cal1.Username != "user" {
		t.Errorf("Unexpected values for first calendar: %+v", cal1)
	}

	cal2 := sec.Calendars[1]
	if cal2.Type != "ical" || cal2.Name != "Holidays" {
		t.Errorf("Unexpected values for second calendar: %+v", cal2)
	}
}

func TestLoadSlideshowConfig(t *testing.T) {
	yamlData := `
slideshow:
  interval: "10s"
  shuffle: true
  transition: "fade"
  target_resolution:
    width: 1920
    height: 1080
  sources:
    - type: "local"
      path: "/tmp/photos"
    - type: "s3"
      bucket: "my-bucket"
      region: "us-east-1"
`
	tmpFile, err := os.CreateTemp("", "slideshow_config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Slideshow.Interval != "10s" {
		t.Errorf("Expected interval 10s, got %s", cfg.Slideshow.Interval)
	}
	if !cfg.Slideshow.Shuffle {
		t.Errorf("Expected shuffle to be true")
	}
	if cfg.Slideshow.Transition != "fade" {
		t.Errorf("Expected transition fade, got %s", cfg.Slideshow.Transition)
	}
	if cfg.Slideshow.TargetResolution.Width != 1920 {
		t.Errorf("Expected width 1920, got %d", cfg.Slideshow.TargetResolution.Width)
	}
	if len(cfg.Slideshow.Sources) != 2 {
		t.Fatalf("Expected 2 sources, got %d", len(cfg.Slideshow.Sources))
	}

	localSrc := cfg.Slideshow.Sources[0]
	if localSrc.Type != "local" || localSrc.Path != "/tmp/photos" {
		t.Errorf("Unexpected values for local source: %+v", localSrc)
	}

	s3Src := cfg.Slideshow.Sources[1]
	if s3Src.Type != "s3" || s3Src.Bucket != "my-bucket" || s3Src.Region != "us-east-1" {
		t.Errorf("Unexpected values for s3 source: %+v", s3Src)
	}
}

func TestLoadUIConfig(t *testing.T) {
	yamlData := `
ui:
  locale: "de-DE"
  time_format: "24h"
  orientation: "portrait"
`
	tmpFile, err := os.CreateTemp("", "ui_config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.UI.Locale != "de-DE" {
		t.Errorf("Expected locale de-DE, got %s", cfg.UI.Locale)
	}
	if cfg.UI.TimeFormat != "24h" {
		t.Errorf("Expected time_format 24h, got %s", cfg.UI.TimeFormat)
	}
	if cfg.UI.Orientation != "portrait" {
		t.Errorf("Expected orientation portrait, got %s", cfg.UI.Orientation)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "ValidConfig",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Sections: []Section{
					{ID: "weather", Region: "center", Type: "weather", Interval: "15m", Weather: &WeatherConfig{}},
				},
			},
			wantErr: false,
		},
		{
			name: "InvalidPort",
			config: Config{
				Server: ServerConfig{Port: 70000},
			},
			wantErr: true,
		},
		{
			name: "InvalidRegion",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Sections: []Section{
					{ID: "news", Region: "nowhere", Type: "rss"},
				},
			},
			wantErr: true,
		},
		{
			name: "WeatherIntervalTooShort",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Sections: []Section{
					{ID: "weather", Region: "center", Type: "weather", Interval: "1m", Weather: &WeatherConfig{}},
				},
			},
			wantErr: true,
		},
		{
			name: "WeatherIntervalInvalidString",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Sections: []Section{
					{ID: "weather", Region: "center", Type: "weather", Interval: "not-a-time", Weather: &WeatherConfig{}},
				},
			},
			wantErr: true,
		},
		{
			name: "RSSIntervalOK",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Sections: []Section{
					{ID: "news", Region: "center", Type: "rss", Interval: "5m"},
				},
			},
			wantErr: false,
		},
		{
			name: "RSSIntervalTooShort",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Sections: []Section{
					{ID: "news", Region: "center", Type: "rss", Interval: "30s"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
