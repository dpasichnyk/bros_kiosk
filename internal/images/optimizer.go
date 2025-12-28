package images

import (
	"image"
	"io"

	"github.com/disintegration/imaging"
)

func Resize(r io.Reader, width, height int) (image.Image, error) {
	img, err := imaging.Decode(r)
	if err != nil {
		return nil, err
	}

	dst := imaging.Fit(img, width, height, imaging.Linear)
	return dst, nil
}
