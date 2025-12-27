package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Slideshow SlideshowConfig `yaml:"slideshow"`
	UI        UIConfig        `yaml:"ui"`
	Sections  []Section       `yaml:"sections"`
}

type ServerConfig struct {
	Port           int    `yaml:"port"`
	Host           string `yaml:"host"`
	UpdateInterval string `yaml:"update_interval"`
}

type UIConfig struct {
	Locale      string `yaml:"locale"`
	TimeFormat  string `yaml:"time_format"` // 12h or 24h
	Orientation string `yaml:"orientation"` // landscape or portrait
}

type SlideshowConfig struct {
	Sources          []SourceConfig `yaml:"sources"`
	Interval         string         `yaml:"interval"`
	Shuffle          bool           `yaml:"shuffle"`
	Transition       string         `yaml:"transition"`
	TargetResolution Resolution     `yaml:"target_resolution"`
}

type SourceConfig struct {
	Type      string `yaml:"type"`
	Path      string `yaml:"path"`       // For local
	Bucket    string `yaml:"bucket"`     // For S3
	Region    string `yaml:"region"`     // For S3
	Prefix    string `yaml:"prefix"`     // For S3
	AccessKey string `yaml:"access_key"` // For S3
	SecretKey string `yaml:"secret_key"` // For S3
	Endpoint  string `yaml:"endpoint"`   // For S3
}

type Resolution struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type Section struct {
	ID        string           `yaml:"id"`
	Region    string           `yaml:"region"`
	Interval  string           `yaml:"interval"`
	Type      string           `yaml:"type"`
	Style     string           `yaml:"style"`
	Weather   *WeatherConfig   `yaml:"weather,omitempty"`
	RSS       *RSSConfig       `yaml:"rss,omitempty"`
	Calendars []CalendarSource `yaml:"calendars,omitempty"`
}

type WeatherConfig struct {
	APIKey  string `yaml:"api_key"`
	City    string `yaml:"city"`
	Units   string `yaml:"units"`
	BaseURL string `yaml:"base_url"`
}

type RSSConfig struct {
	URL string `yaml:"url"`
}

type CalendarSource struct {
	Type     string `yaml:"type"`
	URL      string `yaml:"url"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Color    string `yaml:"color"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	err = yaml.Unmarshal([]byte(expanded), &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	validRegions := map[string]bool{
		"top-left":     true,
		"top-right":    true,
		"center":       true,
		"bottom-left":  true,
		"bottom-right": true,
	}

	for _, s := range c.Sections {
		if s.Region != "" && !validRegions[s.Region] {
			return fmt.Errorf("invalid region '%s' for section '%s'", s.Region, s.ID)
		}

		if s.Interval != "" {
			duration, err := time.ParseDuration(s.Interval)
			if err != nil {
				return fmt.Errorf("invalid interval '%s' for section '%s': %w", s.Interval, s.ID, err)
			}
			// Enforce minimums
			if s.Type == "weather" {
				if duration < 10*time.Minute {
					return fmt.Errorf("interval '%s' for weather section '%s' is too short (minimum 10m)", s.Interval, s.ID)
				}
			} else if duration < 1*time.Minute {
				// Sane minimum for other widgets to prevent spin loops
				return fmt.Errorf("interval '%s' for section '%s' is too short (minimum 1m)", s.Interval, s.ID)
			}
		}

		if s.Type == "weather" && s.Weather != nil {
			if s.Weather.APIKey == "" {
				// We allow empty API key, but maybe we should warn?
				// Actually fetcher handles it gracefully (Setup Required state).
				// So maybe just ensure City is set if we want strictness, but let's keep it loose for now.
			}
		}
	}

	return nil
}
