package server

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"bros_kiosk/internal/images"
	"bros_kiosk/internal/renderer"
	"bros_kiosk/pkg/fetcher"
)

func (s *DashboardServer) ImageHandler(w http.ResponseWriter, r *http.Request) {
	if s.imageRenderer == nil {
		http.Error(w, "Image renderer not configured", http.StatusServiceUnavailable)
		return
	}

	opts := renderer.DefaultOptions()

	if w := r.URL.Query().Get("w"); w != "" {
		if val, err := strconv.Atoi(w); err == nil && val > 0 && val <= 4096 {
			opts.Width = val
		}
	}
	if h := r.URL.Query().Get("h"); h != "" {
		if val, err := strconv.Atoi(h); err == nil && val > 0 && val <= 4096 {
			opts.Height = val
		}
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "png"
	}
	opts.Format = format

	data := s.collectDashboardData(opts.Width, opts.Height)

	img, err := s.imageRenderer.Render(r.Context(), opts, data)
	if err != nil {
		http.Error(w, "Failed to render image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch format {
	case "jpg", "jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, img, &jpeg.Options{Quality: 85})
	case "raw", "rgba":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Image-Width", strconv.Itoa(opts.Width))
		w.Header().Set("X-Image-Height", strconv.Itoa(opts.Height))
		w.Header().Set("X-Image-Format", "RGBA")
		writeRawRGBA(w, img)
	case "rgb565":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Image-Width", strconv.Itoa(opts.Width))
		w.Header().Set("X-Image-Height", strconv.Itoa(opts.Height))
		w.Header().Set("X-Image-Format", "RGB565")
		writeRGB565(w, img)
	default:
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}
}

func (s *DashboardServer) collectDashboardData(targetWidth, targetHeight int) renderer.DashboardData {
	bgImg := s.loadBackgroundImage(targetWidth, targetHeight)

	s.mu.RLock()
	defer s.mu.RUnlock()

	data := renderer.DashboardData{
		Config:      s.config,
		SectionData: make(map[string]interface{}),
		Time:        time.Now(),
		Locale:      s.config.UI.Locale,
		TimeFormat:  s.config.UI.TimeFormat,
		Background:  bgImg,
	}

	for _, sec := range s.config.Sections {
		if result, ok := s.state[sec.ID]; ok && result.Data != nil {
			data.SectionData[sec.ID] = result.Data
		}

		if sec.Type == "weather" && data.Weather == nil {
			if weatherResult, ok := s.state[sec.ID]; ok && weatherResult.Data != nil {
				if wd, ok := weatherResult.Data.(*fetcher.WeatherData); ok {
					data.Weather = &renderer.WeatherData{
						Temp:        wd.Temp,
						Description: wd.Description,
						Icon:        wd.Icon,
						City:        wd.City,
					}
				}
			}
		}

		if sec.Type == "rss" {
			if rssResult, ok := s.state[sec.ID]; ok {
				if rssResult.Data != nil {
					if rd, ok := rssResult.Data.(*fetcher.RSSData); ok {
						for _, item := range rd.Items {
							pubDate, _ := time.Parse(time.RFC1123Z, item.PubDate)
							data.News = append(data.News, renderer.NewsItem{
								Title:   item.Title,
								Summary: item.Summary,
								PubDate: pubDate,
							})
						}
					}
				}
			}
		}

		if sec.Type == "calendar" {
			if calResult, ok := s.state[sec.ID]; ok {
				if calResult.Data != nil {
					if cd, ok := calResult.Data.(*fetcher.CalendarData); ok {
						for _, event := range cd.Events {
							data.Calendar = append(data.Calendar, renderer.CalendarEvent{
								Summary:  event.Summary,
								Start:    event.Start,
								End:      event.End,
								AllDay:   event.AllDay,
								Location: event.Location,
							})
						}
					}
				}
			}
		}
	}

	return data
}

func (s *DashboardServer) loadBackgroundImage(targetWidth, targetHeight int) image.Image {
	if s.scannerMgr == nil {
		return nil
	}

	photos := s.scannerMgr.GetPhotos()
	if len(photos) == 0 {
		return nil
	}

	photoPath := photos[0]
	if len(photos) > 1 {
		idx := int(time.Now().Unix()/30) % len(photos)
		photoPath = photos[idx]
	}

	if cachedPath, found := s.imageCache.Get(photoPath); found {
		f, err := os.Open(cachedPath)
		if err == nil {
			defer f.Close()
			img, _, err := image.Decode(f)
			if err == nil {
				return img
			}
		}
	}

	srcFile, err := os.Open(photoPath)
	if err != nil {
		slog.Debug("Failed to open photo for background", "path", photoPath, "error", err)
		return nil
	}
	defer srcFile.Close()

	resizedImg, err := images.Resize(srcFile, targetWidth, targetHeight)
	if err != nil {
		slog.Debug("Failed to resize background image", "error", err)
		return nil
	}

	s.imageCache.Put(photoPath, resizedImg)

	return resizedImg
}

func writeRawRGBA(w io.Writer, img image.Image) {
	bounds := img.Bounds()
	buf := make([]byte, bounds.Dx()*bounds.Dy()*4)
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			buf[idx] = uint8(r >> 8)
			buf[idx+1] = uint8(g >> 8)
			buf[idx+2] = uint8(b >> 8)
			buf[idx+3] = uint8(a >> 8)
			idx += 4
		}
	}
	w.Write(buf)
}

func writeRGB565(w io.Writer, img image.Image) {
	bounds := img.Bounds()
	buf := make([]byte, bounds.Dx()*bounds.Dy()*2)
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r5 := uint16(r>>11) & 0x1F
			g6 := uint16(g>>10) & 0x3F
			b5 := uint16(b>>11) & 0x1F
			rgb565 := (r5 << 11) | (g6 << 5) | b5
			buf[idx] = uint8(rgb565 & 0xFF)
			buf[idx+1] = uint8(rgb565 >> 8)
			idx += 2
		}
	}
	w.Write(buf)
}
