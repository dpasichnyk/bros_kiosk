package images

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestResize(t *testing.T) {
	// Create 100x100 image
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			src.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}

	img, err := Resize(&buf, 50, 50)
	if err != nil {
		t.Fatalf("Resize failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Expected 50x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
