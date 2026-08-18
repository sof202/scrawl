package main

import (
	"image"
	"image/png"
	"os"
)

// Opens the given file path and encodes the current state of the canvas as a
// png. Images are generally between 1KB-20KB depending on how varied the image
// is (for more information look into the PNG file format).
func (a *ScrawlApp) saveImage(path string) error {
	// Images drawn in scawl only use black (drawn) and white (not-drawn). As
	// such, the image can be purely grayscale. This reduces some complexity
	// and gives a slight speedup.
	img := image.NewGray(image.Rect(0, 0, int(a.width), int(a.height)))

	for i, drawn := range a.canvas {
		var v byte
		if !drawn {
			v = 255 // white
		}
		img.Pix[i] = v
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return err
	}
	return nil
}
