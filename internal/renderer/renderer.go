package renderer

import (
	"context"
	"image"
	"time"
)

type RenderOptions struct {
	Width       int
	Height      int
	Format      string
	ColorDepth  int
	Orientation string
}

type WeatherData struct {
	Temp        float64
	Description string
	Icon        string
	City        string
}

type NewsItem struct {
	Title   string
	Summary string
	PubDate time.Time
}

type CalendarEvent struct {
	Summary  string
	Start    time.Time
	End      time.Time
	AllDay   bool
	Location string
}

type DashboardData struct {
	Config interface{}

	SectionData map[string]interface{}

	Time       time.Time
	Locale     string
	TimeFormat string
	Weather    *WeatherData
	News       []NewsItem
	Calendar   []CalendarEvent
	Background image.Image
}

type Renderer interface {
	Render(ctx context.Context, opts RenderOptions, data DashboardData) (image.Image, error)
	Name() string
}

func DefaultOptions() RenderOptions {
	return RenderOptions{
		Width:       1920,
		Height:      1080,
		Format:      "png",
		ColorDepth:  24,
		Orientation: "landscape",
	}
}
