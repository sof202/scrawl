package main

import (
	"image/color"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	width       int32 = 800
	height      int32 = 600
	bytesPerRow int32 = 4 // len("RGBA") = 4
	stride            = int(width * bytesPerRow)
)

var (
	white = color.RGBA{255, 255, 255, 255}

	// The idea here is that, as we don't care about colours (only black
	// strokes on a white background), our canvas is just a vector of
	// black/white (Boolean). We then can just convert the canvas to pixels on
	// each frame.
	canvas = make([]bool, width*height) // false -> white, true -> black
	pixels = make([]byte, width*height*bytesPerRow)
)

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		"scrawl",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		width,
		height,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(
		window,
		-1,
		sdl.RENDERER_ACCELERATED,
	)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(
		uint32(sdl.PIXELFORMAT_RGBA32), //
		sdl.TEXTUREACCESS_STREAMING,
		width,
		height,
	)
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	update := func() {
		canvasToPixels()
		texture.Update(
			nil,
			unsafe.Pointer(unsafe.SliceData(pixels)),
			stride,
		)
		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}
		update()
		sdl.Delay(16)
	}
}

func canvasToPixels() {
	for i, drawn := range canvas {
		idx := i * int(bytesPerRow)
		var v byte
		if !drawn {
			v = 255 // white
		}
		pixels[idx+0] = v
		pixels[idx+1] = v
		pixels[idx+2] = v
		pixels[idx+3] = 255 // always no alpha/transparency
	}

}
