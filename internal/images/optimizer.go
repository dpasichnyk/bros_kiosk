package images

import (
	"image"
	"io"

	"github.com/disintegration/imaging"
)

// Resize decodes an image from reader and resizes it to fit within width/height, maintaining aspect ratio.
// It uses linear interpolation for performance suitable for Raspberry Pi.
func Resize(r io.Reader, width, height int) (image.Image, error) {
	img, err := imaging.Decode(r)
	if err != nil {
		return nil, err
	}

	dst := imaging.Fit(img, width, height, imaging.Linear)
	return dst, nil
}
