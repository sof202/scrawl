package main

import (
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type ScrawlApp struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture

	width, height int32
	stride        int

	// The idea here is that, as we don't care about colours (only black
	// strokes on a white background), our canvas is just a vector of
	// black/white (Boolean). We then can just convert the canvas to pixels on
	// each frame.
	canvas []bool
	pixels []byte

	drawing                                bool
	brushSize                              int32
	prevMouseX, prevMouseY, mouseX, mouseY int32
}

func NewApp(width, height int32) (*ScrawlApp, error) {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return &ScrawlApp{}, err
	}

	window, err := sdl.CreateWindow(
		"scrawl",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		width,
		height,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return &ScrawlApp{}, err
	}

	renderer, err := sdl.CreateRenderer(
		window,
		-1,
		sdl.RENDERER_ACCELERATED,
	)
	if err != nil {
		return &ScrawlApp{}, err
	}

	texture, err := renderer.CreateTexture(
		uint32(sdl.PIXELFORMAT_RGBA32), //
		sdl.TEXTUREACCESS_STREAMING,
		width,
		height,
	)
	if err != nil {
		return &ScrawlApp{}, err
	}

	return &ScrawlApp{
		window:     window,
		renderer:   renderer,
		texture:    texture,
		width:      width,
		height:     height,
		stride:     int(width * bytesPerRow),
		canvas:     make([]bool, width*height),
		pixels:     make([]byte, width*height*bytesPerRow),
		brushSize:  int32(10),
		prevMouseX: width / 2,
		prevMouseY: height / 2,
		mouseX:     width / 2,
		mouseY:     height / 2,
	}, nil
}

func (a *ScrawlApp) Close() {
	a.texture.Destroy()
	a.renderer.Destroy()
	a.window.Destroy()
	sdl.Quit()
}

func (a *ScrawlApp) Run(outputPath string) error {
	updateScreen := func() {
		a.canvasToPixels()
		a.texture.Update(
			nil,
			unsafe.Pointer(unsafe.SliceData(a.pixels)),
			a.stride,
		)
		a.renderer.Copy(a.texture, nil, nil)

		// Must come after copying texture as the texture covers the entire
		// window.
		a.drawRing(a.mouseX, a.mouseY, a.brushSize)
		a.renderer.Present()
	}

	// Event listener
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				a.saveImage(outputPath)

			case *sdl.KeyboardEvent:
				if e.Type != sdl.KEYDOWN {
					continue
				}
				switch e.Keysym.Sym {
				case sdl.K_c: // clear
					a.canvas = make([]bool, a.width*a.height)
				case sdl.K_ESCAPE: // exit without saving
					running = false
				}

			case *sdl.MouseWheelEvent:
				a.brushSize += e.Y * 4
				if a.brushSize > maxBrushSize {
					a.brushSize = maxBrushSize
				}
				if a.brushSize < minBrushSize {
					a.brushSize = minBrushSize
				}

			case *sdl.MouseButtonEvent:
				if e.Button != sdl.BUTTON_LEFT {
					continue
				}
				switch e.Type {
				case sdl.MOUSEBUTTONDOWN:
					a.prevMouseX, a.prevMouseY = e.X, e.Y
					a.drawing = true

					// Accounts for the case where user only clicks the mouse.
					// In such cases the line drawing algorithm might not proc
					// as no mouse motion is detected.
					a.drawCircle(a.prevMouseX, a.prevMouseY, a.brushSize)
				case sdl.MOUSEBUTTONUP:
					a.drawing = false
				}

			case *sdl.MouseMotionEvent:
				a.mouseX, a.mouseY = e.X, e.Y
				if a.drawing {
					a.drawLine(a.prevMouseX, a.prevMouseY, a.mouseX, a.mouseY)
					a.prevMouseX, a.prevMouseY = e.X, e.Y
				}
			}
		}
		updateScreen()
		sdl.Delay(16)
	}
	return nil
}
