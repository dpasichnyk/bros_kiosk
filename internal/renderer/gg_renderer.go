package renderer

import (
	"context"
	"embed"
	"fmt"
	"image"
	"image/color"
	"time"

	"bros_kiosk/internal/config"
	"bros_kiosk/pkg/fetcher"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/goodsign/monday"
	"golang.org/x/image/font"
)

//go:embed fonts/Roboto-Regular.ttf fonts/Roboto-Light.ttf
var embeddedFonts embed.FS

type GGRenderer struct {
	fontRegular *truetype.Font
	fontLight   *truetype.Font
}

func NewGGRenderer() (*GGRenderer, error) {
	regBytes, err := embeddedFonts.ReadFile("fonts/Roboto-Regular.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded font: %w", err)
	}
	fReg, err := truetype.Parse(regBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse regular font: %w", err)
	}

	lightBytes, err := embeddedFonts.ReadFile("fonts/Roboto-Light.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded font: %w", err)
	}
	fLight, err := truetype.Parse(lightBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse light font: %w", err)
	}

	return &GGRenderer{
		fontRegular: fReg,
		fontLight:   fLight,
	}, nil
}

func (r *GGRenderer) Name() string {
	return "gg"
}

func (r *GGRenderer) Render(ctx context.Context, opts RenderOptions, data DashboardData) (image.Image, error) {
	dc := gg.NewContext(opts.Width, opts.Height)

	r.drawBackground(dc, opts, data)

	clockHeight := r.drawClock(dc, opts, data)

	cfg, ok := data.Config.(*config.Config)
	if !ok {
		return dc.Image(), nil
	}

	padding := float64(opts.Width) * 0.025

	colWidth := float64(opts.Width) * 0.30

	yOffsets := map[string]float64{
		"top-left":     padding * 1.5,
		"top-right":    padding * 1.5,
		"center":       clockHeight + padding*2,
		"bottom-left":  float64(opts.Height) * 0.6,
		"bottom-right": float64(opts.Height) * 0.6,
	}

	xPos := map[string]float64{
		"top-left":     padding,
		"bottom-left":  padding,
		"center":       (float64(opts.Width) - colWidth) / 2,
		"top-right":    float64(opts.Width) - padding - colWidth,
		"bottom-right": float64(opts.Width) - padding - colWidth,
	}

	var locale monday.Locale = monday.LocaleEnUS
	if data.Locale != "" {
		locale = monday.Locale(data.Locale)
	}

	for _, sec := range cfg.Sections {
		region := sec.Region
		if region == "" {
			region = "center"
		}

		x, okX := xPos[region]
		y, okY := yOffsets[region]

		if !okX || !okY {
			continue
		}

		var heightDrawn float64

		secData, hasData := data.SectionData[sec.ID]

		switch sec.Type {
		case "weather":
			if hasData {
				if wd, ok := secData.(*fetcher.WeatherData); ok {
					heightDrawn = r.drawWeather(dc, opts, x, y, colWidth, wd)
				}
			} else if data.Weather != nil {
				heightDrawn = r.drawWeather(dc, opts, x, y, colWidth, &fetcher.WeatherData{
					Temp: data.Weather.Temp, Description: data.Weather.Description, City: data.Weather.City, Icon: data.Weather.Icon,
				})
			}
		case "rss":
			if hasData {
				if rd, ok := secData.(*fetcher.RSSData); ok {
					heightDrawn = r.drawRSS(dc, opts, x, y, colWidth, rd, locale)
				}
			}
		case "calendar":
			if hasData {
				if cd, ok := secData.(*fetcher.CalendarData); ok {
					heightDrawn = r.drawCalendar(dc, opts, x, y, colWidth, cd, locale)
				}
			}
		}

		if heightDrawn > 0 {
			yOffsets[region] += heightDrawn + padding
		}
	}

	return dc.Image(), nil
}

func (r *GGRenderer) fontFace(size float64, light bool) font.Face {
	f := r.fontRegular
	if light {
		f = r.fontLight
	}
	return truetype.NewFace(f, &truetype.Options{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func (r *GGRenderer) drawBackground(dc *gg.Context, opts RenderOptions, data DashboardData) {
	if data.Background != nil {
		bgBounds := data.Background.Bounds()
		scaleX := float64(opts.Width) / float64(bgBounds.Dx())
		scaleY := float64(opts.Height) / float64(bgBounds.Dy())
		scale := scaleX
		if scaleY > scale {
			scale = scaleY
		}

		newW := int(float64(bgBounds.Dx()) * scale)
		newH := int(float64(bgBounds.Dy()) * scale)
		offsetX := (opts.Width - newW) / 2
		offsetY := (opts.Height - newH) / 2

		dc.Push()
		dc.Translate(float64(offsetX), float64(offsetY))
		dc.Scale(scale, scale)
		dc.DrawImage(data.Background, 0, 0)
		dc.Pop()

		dc.SetRGBA(0, 0, 0, 0.4)
		dc.DrawRectangle(0, 0, float64(opts.Width), float64(opts.Height))
		dc.Fill()
	} else {
		dc.SetColor(color.Black)
		dc.Clear()
	}
}

func (r *GGRenderer) drawClock(dc *gg.Context, opts RenderOptions, data DashboardData) float64 {
	centerX := float64(opts.Width) / 2
	clockY := float64(opts.Height) * 0.13

	timeFontSize := float64(opts.Height) * 0.12
	dateFontSize := float64(opts.Height) * 0.032

	dc.SetFontFace(r.fontFace(timeFontSize, true))
	dc.SetColor(color.White)

	t := data.Time
	if t.IsZero() {
		t = time.Now()
	}

	var timeStr string
	if data.TimeFormat == "12h" {
		timeStr = t.Format("3:04")
	} else {
		timeStr = t.Format("15:04")
	}

	dc.DrawStringAnchored(timeStr, centerX, clockY, 0.5, 0.5)

	dc.SetFontFace(r.fontFace(dateFontSize, true))
	dc.SetRGBA(1, 1, 1, 0.7)

	var locale monday.Locale = monday.LocaleEnUS
	if data.Locale != "" {
		locale = monday.Locale(data.Locale)
	}
	dateStr := monday.Format(t, "Monday, January 2", locale)

	dc.DrawStringAnchored(dateStr, centerX, clockY+timeFontSize*0.7, 0.5, 0.5)

	return clockY + timeFontSize*0.7 + dateFontSize
}

func (r *GGRenderer) drawWeather(dc *gg.Context, opts RenderOptions, x, y, width float64, data *fetcher.WeatherData) float64 {
	centerX := x + width/2

	if data == nil {
		return 0
	}

	if data.Description == "Setup Required" {
		setupSize := float64(opts.Height) * 0.022
		dc.SetFontFace(r.fontFace(setupSize, true))
		dc.SetRGBA(1, 1, 1, 0.5)
		dc.DrawStringAnchored("Setup Required", centerX, y, 0.5, 0.0)
		return setupSize * 1.5
	}

	tempFontSize := float64(opts.Height) * 0.07
	condFontSize := float64(opts.Height) * 0.022

	curY := y + tempFontSize/2

	dc.SetFontFace(r.fontFace(tempFontSize, true))
	dc.SetColor(color.White)
	tempStr := fmt.Sprintf("%.0f°", data.Temp)
	dc.DrawStringAnchored(tempStr, centerX, curY, 0.5, 0.5)

	curY += tempFontSize * 0.6

	dc.SetFontFace(r.fontFace(condFontSize, true))
	dc.SetRGBA(1, 1, 1, 0.7)
	dc.DrawStringAnchored(data.Description, centerX, curY, 0.5, 0.5)

	curY += tempFontSize * 0.4

	dc.SetFontFace(r.fontFace(condFontSize*0.8, true))
	dc.SetRGBA(1, 1, 1, 0.5)
	dc.DrawStringAnchored(data.City, centerX, curY, 0.5, 0.5)

	return (curY - y) + condFontSize
}

func (r *GGRenderer) drawRSS(dc *gg.Context, opts RenderOptions, x, y, width float64, data *fetcher.RSSData, locale monday.Locale) float64 {
	if len(data.Items) == 0 {
		return 0
	}

	startY := y
	headerSize := float64(opts.Height) * 0.014
	titleSize := float64(opts.Height) * 0.020
	summarySize := float64(opts.Height) * 0.016
	timeSize := float64(opts.Height) * 0.013

	dc.SetFontFace(r.fontFace(headerSize, false))
	dc.SetRGBA(1, 1, 1, 0.45)
	dc.DrawString("NEWS", x, y+headerSize)
	y += headerSize * 3

	maxItems := 5
	if len(data.Items) < maxItems {
		maxItems = len(data.Items)
	}

	for i := 0; i < maxItems; i++ {
		item := data.Items[i]

		dc.SetFontFace(r.fontFace(titleSize, false))
		dc.SetColor(color.White)
		lines := dc.WordWrap(item.Title, width)
		if len(lines) > 2 {
			lines = lines[:2]
		}
		for _, line := range lines {
			dc.DrawString(line, x, y)
			y += titleSize * 1.3
		}

		if item.Summary != "" {
			dc.SetFontFace(r.fontFace(summarySize, true))
			dc.SetRGBA(1, 1, 1, 0.65)
			summaryLines := dc.WordWrap(item.Summary, width)
			if len(summaryLines) > 2 {
				summaryLines = summaryLines[:2]
			}
			for _, line := range summaryLines {
				dc.DrawString(line, x, y)
				y += summarySize * 1.25
			}
		}

		dc.SetFontFace(r.fontFace(timeSize, true))
		dc.SetRGBA(1, 1, 1, 0.45)

		pubDate, _ := time.Parse(time.RFC1123Z, item.PubDate)
		timeAgo := formatRelativeTime(pubDate)
		dc.DrawString(timeAgo, x, y)
		y += timeSize * 1.5

		y += titleSize * 1.0
	}

	return y - startY
}

func (r *GGRenderer) drawCalendar(dc *gg.Context, opts RenderOptions, x, y, width float64, data *fetcher.CalendarData, locale monday.Locale) float64 {
	if len(data.Events) == 0 {
		return 0
	}

	startY := y
	headerSize := float64(opts.Height) * 0.012
	dateSize := float64(opts.Height) * 0.022
	timeSize := float64(opts.Height) * 0.014
	titleSize := float64(opts.Height) * 0.022
	locationSize := float64(opts.Height) * 0.016

	badgeWidth := float64(opts.Width) * 0.07
	badgeHeight := float64(opts.Height) * 0.07
	badgePadding := float64(opts.Width) * 0.012
	padding := float64(opts.Width) * 0.025
	cornerRadius := 4.0

	dc.SetFontFace(r.fontFace(headerSize, false))
	dc.SetRGBA(1, 1, 1, 0.45)

	headerWidth, _ := dc.MeasureString("CALENDAR")
	dc.DrawString("CALENDAR", x+width-headerWidth, y+headerSize)
	y += headerSize * 3

	maxItems := 5
	if len(data.Events) < maxItems {
		maxItems = len(data.Events)
	}

	for i := 0; i < maxItems; i++ {
		event := data.Events[i]

		badgeY := y

		maxTitleWidth := width * 0.6
		totalWidth := badgeWidth + badgePadding + maxTitleWidth

		groupX := x + width - totalWidth

		badgeX := groupX
		textXLocal := badgeX + badgeWidth + badgePadding

		dc.SetRGBA(1, 1, 1, 0.15)
		dc.DrawRoundedRectangle(badgeX, badgeY, badgeWidth, badgeHeight, cornerRadius)
		dc.Fill()

		dc.SetFontFace(r.fontFace(dateSize, false))
		dc.SetColor(color.White)

		dateStr := monday.Format(event.Start, "Jan 2", locale)
		dc.DrawStringAnchored(dateStr, badgeX+badgeWidth/2, badgeY+badgeHeight*0.45, 0.5, 0.5)

		dc.SetFontFace(r.fontFace(timeSize, true))
		dc.SetRGBA(1, 1, 1, 0.7)
		var timeStr string
		if event.AllDay {
			timeStr = "All Day"
		} else {
			timeStr = event.Start.Format("15:04")
		}
		dc.DrawStringAnchored(timeStr, badgeX+badgeWidth/2, badgeY+badgeHeight*0.75, 0.5, 0.5)

		dc.SetFontFace(r.fontFace(titleSize, false))
		dc.SetColor(color.White)

		title := event.Summary
		lines := dc.WordWrap(title, maxTitleWidth)
		if len(lines) > 2 {
			lines = lines[:2]
		}

		currentTextY := badgeY + badgeHeight*0.45 + titleSize*0.35
		for _, line := range lines {
			dc.DrawString(line, textXLocal, currentTextY)
			currentTextY += titleSize * 1.2
		}

		if event.Location != "" {
			dc.SetFontFace(r.fontFace(locationSize, true))
			dc.SetRGBA(1, 1, 1, 0.6)
			locY := badgeY + badgeHeight*0.75 + locationSize*0.35
			if currentTextY > locY-locationSize {
				locY = currentTextY + locationSize*0.2
			}
			dc.DrawString(event.Location, textXLocal, locY)
		}

		y += badgeHeight + padding*0.5
	}

	return y - startY
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	return t.Format("Jan 2")
}
