package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

const (
	width  int32 = 800
	height int32 = 600
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

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}
		renderer.Copy(texture, nil, nil)
		renderer.Present()
		sdl.Delay(16)
	}
}
