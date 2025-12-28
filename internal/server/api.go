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

func (s *DashboardServer) PhotosListHandler(w http.ResponseWriter, r *http.Request) {
	photos := s.scannerMgr.GetPhotos()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"photos": photos,
	})
}

func (s *DashboardServer) AssetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	rawPath := strings.TrimPrefix(r.URL.Path, "/assets/photos/")
	if rawPath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	filePath, err := url.QueryUnescape(rawPath)
	if err != nil {
		http.Error(w, "Invalid path encoding", http.StatusBadRequest)
		return
	}

	if cachedPath, found := s.imageCache.Get(filePath); found {
		http.ServeFile(w, r, cachedPath)
		return
	}

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
	}
	if targetH == 0 {
		targetH = 1080
	}

	resizedImg, err := images.Resize(srcFile, targetW, targetH)
	if err != nil {
		http.Error(w, "Failed to resize image", http.StatusInternalServerError)
		return
	}

	savedPath, err := s.imageCache.Put(filePath, resizedImg)
	if err != nil {
		http.Error(w, "Failed to save to cache", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, savedPath)
}

func (s *DashboardServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	clientHash := r.Header.Get("X-Dashboard-Hash")

	s.mu.RLock()
	updates := make(map[string]interface{})
	for _, sec := range s.config.Sections {
		if res, ok := s.state[sec.ID]; ok {
			updates[sec.ID] = res
		} else if sec.Type == "rss" {
			if res, ok := s.state["rss"]; ok {
				updates[sec.ID] = res
			}
		} else if sec.Type == "weather" {
			if res, ok := s.state["weather"]; ok {
				updates[sec.ID] = res
			}
		}
	}
	s.mu.RUnlock()

	fullHash, err := hashing.Hash(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if clientHash == fullHash {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	response := map[string]interface{}{
		"status":  "ok",
		"hash":    fullHash,
		"updates": updates,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
