package server

import (
	"encoding/json"
	"net/http"

	"bros_kiosk/internal/hashing"
	"bros_kiosk/internal/images"
	"net/url"
	"os"
	"strings"
)

// PhotosListHandler returns the list of available photos
func (s *DashboardServer) PhotosListHandler(w http.ResponseWriter, r *http.Request) {
	photos := s.scannerMgr.GetPhotos()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"photos": photos,
	})
}

// AssetHandler serves optimized images
func (s *DashboardServer) AssetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	// Extract file path from URL
	rawPath := strings.TrimPrefix(r.URL.Path, "/assets/photos/")
	if rawPath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	// URL decode the path
	filePath, err := url.QueryUnescape(rawPath)
	if err != nil {
		http.Error(w, "Invalid path encoding", http.StatusBadRequest)
		return
	}

	// 1. Check Cache
	if cachedPath, found := s.imageCache.Get(filePath); found {
		http.ServeFile(w, r, cachedPath)
		return
	}

	// 2. If not in cache, optimize and cache
	// Check if it's a local file (basic check)
	// In a real S3 scenario, we'd need to identify source type from the path or a map
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	srcFile, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open source image", http.StatusInternalServerError)
		return
	}
	defer srcFile.Close()

	targetW := s.config.Slideshow.TargetResolution.Width
	targetH := s.config.Slideshow.TargetResolution.Height
	if targetW == 0 {
		targetW = 1920
	} // Default
	if targetH == 0 {
		targetH = 1080
	} // Default

	resizedImg, err := images.Resize(srcFile, targetW, targetH)
	if err != nil {
		http.Error(w, "Failed to resize image", http.StatusInternalServerError)
		return
	}

	// 3. Save to Cache
	savedPath, err := s.imageCache.Put(filePath, resizedImg)
	if err != nil {
		// Log error but try to serve directly? Or fail?
		// For now, fail
		http.Error(w, "Failed to save to cache", http.StatusInternalServerError)
		return
	}

	// 4. Serve from Cache
	http.ServeFile(w, r, savedPath)
}

// UpdateHandler returns the latest updates for the dashboard
func (s *DashboardServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	clientHash := r.Header.Get("X-Dashboard-Hash")

	// 1. Get current state from FetcherManager results
	s.mu.RLock()
	// Map internal state keys to dashboard section IDs
	updates := make(map[string]interface{})
	for _, sec := range s.config.Sections {
		// Try to find by section ID (preferred)
		if res, ok := s.state[sec.ID]; ok {
			updates[sec.ID] = res
		} else if sec.Type == "rss" {
			// Fallback for rss if keyed as "rss" internally
			if res, ok := s.state["rss"]; ok {
				updates[sec.ID] = res
			}
		} else if sec.Type == "weather" {
			// Fallback for weather if keyed as "weather" internally
			if res, ok := s.state["weather"]; ok {
				updates[sec.ID] = res
			}
		}
	}
	s.mu.RUnlock()

	// 2. Calculate current full hash
	fullHash, err := hashing.Hash(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Compare with client hash
	if clientHash == fullHash {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// 5. Build response
	response := map[string]interface{}{
		"status":  "ok",
		"hash":    fullHash,
		"updates": updates,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
