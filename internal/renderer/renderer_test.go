package renderer

import (
	"context"
	"image"
	"testing"
	"time"
)

func TestNewGGRenderer(t *testing.T) {
	r, err := NewGGRenderer()
	if err != nil {
		t.Fatalf("NewGGRenderer() error = %v", err)
	}
	if r == nil {
		t.Fatalf("NewGGRenderer() returned nil")
	}
	if r.Name() != "gg" {
		t.Errorf("Name() = %v, want gg", r.Name())
	}
}

func TestGGRenderer_Render_Basic(t *testing.T) {
	r, err := NewGGRenderer()
	if err != nil {
		t.Fatalf("NewGGRenderer() error = %v", err)
	}

	opts := RenderOptions{
		Width:       640,
		Height:      480,
		Format:      "png",
		Orientation: "landscape",
	}

	data := DashboardData{
		Time:       time.Date(2024, 12, 25, 14, 30, 0, 0, time.UTC),
		Locale:     "en-US",
		TimeFormat: "24h",
	}

	img, err := r.Render(context.Background(), opts, data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if img == nil {
		t.Fatalf("Render() returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != opts.Width || bounds.Dy() != opts.Height {
		t.Errorf("Image size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), opts.Width, opts.Height)
	}
}

func TestGGRenderer_Render_WithWeather(t *testing.T) {
	r, err := NewGGRenderer()
	if err != nil {
		t.Fatalf("NewGGRenderer() error = %v", err)
	}

	opts := RenderOptions{
		Width:  800,
		Height: 600,
	}

	data := DashboardData{
		Time:       time.Now(),
		TimeFormat: "12h",
		Weather: &WeatherData{
			Temp:        22.5,
			Description: "Partly Cloudy",
			Icon:        "02d",
			City:        "Berlin",
		},
	}

	img, err := r.Render(context.Background(), opts, data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if img == nil {
		t.Fatalf("Render() returned nil image")
	}
}

func TestGGRenderer_Render_WithNews(t *testing.T) {
	r, err := NewGGRenderer()
	if err != nil {
		t.Fatalf("NewGGRenderer() error = %v", err)
	}

	data := DashboardData{
		Time: time.Now(),
		News: []NewsItem{
			{Title: "Test Headline 1", Summary: "Brief summary", PubDate: time.Now()},
			{Title: "Test Headline 2", Summary: "Another summary", PubDate: time.Now()},
		},
	}

	img, err := r.Render(context.Background(), DefaultOptions(), data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if img == nil {
		t.Fatalf("Render() returned nil image")
	}
}

func TestGGRenderer_Render_WithCalendar(t *testing.T) {
	r, err := NewGGRenderer()
	if err != nil {
		t.Fatalf("NewGGRenderer() error = %v", err)
	}

	data := DashboardData{
		Time: time.Now(),
		Calendar: []CalendarEvent{
			{Summary: "Team Meeting", Start: time.Now().Add(2 * time.Hour), AllDay: false},
			{Summary: "Holiday", Start: time.Now().Add(24 * time.Hour), AllDay: true, Location: "Office"},
		},
	}

	img, err := r.Render(context.Background(), DefaultOptions(), data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if img == nil {
		t.Fatalf("Render() returned nil image")
	}
}

func TestGGRenderer_Render_WithBackground(t *testing.T) {
	r, err := NewGGRenderer()
	if err != nil {
		t.Fatalf("NewGGRenderer() error = %v", err)
	}

	bg := image.NewRGBA(image.Rect(0, 0, 1920, 1080))

	data := DashboardData{
		Time:       time.Now(),
		Background: bg,
	}

	img, err := r.Render(context.Background(), DefaultOptions(), data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if img == nil {
		t.Fatalf("Render() returned nil image")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Width != 1920 {
		t.Errorf("Width = %d, want 1920", opts.Width)
	}
	if opts.Height != 1080 {
		t.Errorf("Height = %d, want 1080", opts.Height)
	}
	if opts.Format != "png" {
		t.Errorf("Format = %s, want png", opts.Format)
	}
}
