package images

import (
	"image"
	"os"
	"testing"
)

func TestDiskCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "img_cache")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache, err := NewDiskCache(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	key := "test_image_1"

	// Get - Miss
	_, found := cache.Get(key)
	if found {
		t.Error("Expected miss, got hit")
	}

	// Put
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	path, err := cache.Put(key, img)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("File not created")
	}

	// Get - Hit
	cachedPath, found := cache.Get(key)
	if !found {
		t.Error("Expected hit, got miss")
	}
	if cachedPath != path {
		t.Errorf("Expected path %s, got %s", path, cachedPath)
	}
}
