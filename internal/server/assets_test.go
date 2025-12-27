package server

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"bros_kiosk/internal/config"
	"bros_kiosk/internal/images"
	"bros_kiosk/internal/scanner"
)

func createTestImage(t *testing.T, path string) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestAssetHandler(t *testing.T) {
	// 1. Setup Temp Dirs
	tmpSrc, err := os.MkdirTemp("", "src_photos")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpSrc)

	tmpCache, err := os.MkdirTemp("", "cache_photos")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpCache)

	// 2. Create Test Image
	imgName := "test.png"
	imgPath := filepath.Join(tmpSrc, imgName)
	createTestImage(t, imgPath)

	// 3. Configure Server
	cfg := &config.Config{
		Slideshow: config.SlideshowConfig{
			TargetResolution: config.Resolution{Width: 50, Height: 50},
		},
	}
	
	cache, err := images.NewDiskCache(tmpCache)
	if err != nil {
		t.Fatal(err)
	}
	
	scanMgr := scanner.NewManager() // Empty is fine, handler bypasses it for file access

	srv := &DashboardServer{
		config:     cfg,
		imageCache: cache,
		scannerMgr: scanMgr,
	}

	// 4. Create Request
	// Encode path
	encodedPath := url.QueryEscape(imgPath)
	req := httptest.NewRequest("GET", "/assets/photos/"+encodedPath, nil)
	w := httptest.NewRecorder()

	// 5. Serve
	srv.AssetHandler(w, req)

	// 6. Verify
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify Cache File Exists
	// We don't know the hash filename easily here without importing internal logic, 
	// but we can check if directory is not empty
	entries, _ := os.ReadDir(tmpCache)
	if len(entries) == 0 {
		t.Error("Expected cache entry to be created")
	}
}

func TestPhotosListHandler(t *testing.T) {
	scanMgr := scanner.NewManager()
	// We can't easily inject mock scanners into Manager from here as Scanner is interface in internal/scanner
	// and we are in internal/server. But we can trust Manager logic tested elsewhere.
	// We just want to ensure JSON response structure.
	
	srv := &DashboardServer{
		scannerMgr: scanMgr,
	}

	req := httptest.NewRequest("GET", "/api/photos", nil)
	w := httptest.NewRecorder()

	srv.PhotosListHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}
