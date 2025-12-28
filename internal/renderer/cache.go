package renderer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"sync"
	"time"
)

type CachedRenderer struct {
	delegate Renderer
	cache    map[string]*cacheEntry
	mu       sync.RWMutex
	ttl      time.Duration
	maxSize  int
}

type cacheEntry struct {
	image     image.Image
	createdAt time.Time
}

func NewCachedRenderer(delegate Renderer, ttl time.Duration) *CachedRenderer {
	return &CachedRenderer{
		delegate: delegate,
		cache:    make(map[string]*cacheEntry),
		ttl:      ttl,
		maxSize:  10,
	}
}

func (r *CachedRenderer) Name() string {
	return "cached-" + r.delegate.Name()
}

func (r *CachedRenderer) Render(ctx context.Context, opts RenderOptions, data DashboardData) (image.Image, error) {
	key := r.cacheKey(opts, data)

	r.mu.RLock()
	if entry, ok := r.cache[key]; ok {
		if time.Since(entry.createdAt) < r.ttl {
			r.mu.RUnlock()
			return entry.image, nil
		}
	}
	r.mu.RUnlock()

	img, err := r.delegate.Render(ctx, opts, data)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	if len(r.cache) >= r.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range r.cache {
			if oldestKey == "" || v.createdAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.createdAt
			}
		}
		if oldestKey != "" {
			delete(r.cache, oldestKey)
		}
	}
	r.cache[key] = &cacheEntry{
		image:     img,
		createdAt: time.Now(),
	}
	r.mu.Unlock()

	return img, nil
}

func (r *CachedRenderer) cacheKey(opts RenderOptions, data DashboardData) string {
	h := sha256.New()

	keyData := struct {
		Width      int
		Height     int
		Format     string
		TimeMinute string
		Weather    *WeatherData
		NewsLen    int
		CalLen     int
	}{
		Width:      opts.Width,
		Height:     opts.Height,
		Format:     opts.Format,
		TimeMinute: data.Time.Truncate(time.Minute).Format(time.RFC3339),
		Weather:    data.Weather,
		NewsLen:    len(data.News),
		CalLen:     len(data.Calendar),
	}

	json.NewEncoder(h).Encode(keyData)
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (r *CachedRenderer) ClearCache() {
	r.mu.Lock()
	r.cache = make(map[string]*cacheEntry)
	r.mu.Unlock()
}

func imageToBytes(img image.Image) []byte {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
