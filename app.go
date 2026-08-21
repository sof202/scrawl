package main

import (
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type ScrawlApp struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	running  bool

	width, height int32
	stride        int

	// The idea here is that, as we don't care about colours (only black
	// strokes on a white background), our canvas is just a vector of
	// black/white (Boolean). We then can just convert the canvas to pixels on
	// each frame.
	canvas []bool
	pixels []byte

	// A single app is destined to be saved in only a single location, as such
	// the app owns it's output path. This isn't common in other similar
	// programs, but eliminates it as a pass-through variable and is
	// functionally equivalent.
	outputPath string

	drawing                                bool
	brushSize                              int32
	prevMouseX, prevMouseY, mouseX, mouseY int32
}

func NewApp(cli CLI) (*ScrawlApp, error) {
	width := cli.width
	height := cli.height

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	success := false
	defer func() {
		if !success {
			sdl.Quit()
		}
	}()

	// Rather than use sdl.WINDOWPOS_CENTERED (which has some edge case bugs),
	// we compute the centre of the display manually.
	displayIndex := 0
	bounds, err := sdl.GetDisplayBounds(displayIndex)
	if err != nil {
		return nil, err
	}
	xPosition := bounds.X + (bounds.W-width)/2
	yPosition := bounds.Y + (bounds.H-height)/2

	window, err := sdl.CreateWindow(
		"scrawl",
		xPosition,
		yPosition,
		width,
		height,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			window.Destroy()
		}
	}()

	renderer, err := sdl.CreateRenderer(
		window,
		-1,
		sdl.RENDERER_ACCELERATED,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			renderer.Destroy()
		}
	}()

	texture, err := renderer.CreateTexture(
		uint32(sdl.PIXELFORMAT_RGBA32), //
		sdl.TEXTUREACCESS_STREAMING,
		width,
		height,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			texture.Destroy()
		}
	}()

	success = true
	return &ScrawlApp{
		window:     window,
		renderer:   renderer,
		texture:    texture,
		running:    true,
		width:      width,
		height:     height,
		stride:     int(width * bytesPerRow),
		canvas:     make([]bool, width*height),
		pixels:     make([]byte, width*height*bytesPerRow),
		outputPath: cli.outputPath,
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

func (a *ScrawlApp) Run() error {
	for a.running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if err := a.handleEvent(event); err != nil {
				return err
			}
		}
		a.updateScreen()
		sdl.Delay(16) // Let the CPU breathe
	}
	return nil
}

func (a *ScrawlApp) handleEvent(event sdl.Event) error {
	switch e := event.(type) {
	case *sdl.QuitEvent:
		a.running = false
		return a.saveImage()
	case *sdl.KeyboardEvent:
		a.handleKeyboardEvent(e)
	case *sdl.MouseWheelEvent:
		a.handleMouseWheelEvent(e)
	case *sdl.MouseButtonEvent:
		a.handleMouseButtonEvent(e)
	case *sdl.MouseMotionEvent:
		a.handleMouseMotionEvent(e)
	}
	return nil
}

func (a *ScrawlApp) handleKeyboardEvent(e *sdl.KeyboardEvent) {
	if e.Type != sdl.KEYDOWN {
		return
	}
	switch e.Keysym.Sym {
	case sdl.K_c: // clear
		a.canvas = make([]bool, a.width*a.height)
	case sdl.K_ESCAPE: // exit without saving
		a.running = false
	}
}

func (a *ScrawlApp) handleMouseWheelEvent(e *sdl.MouseWheelEvent) {
	a.brushSize += e.Y * 4
	if a.brushSize > maxBrushSize {
		a.brushSize = maxBrushSize
	}
	if a.brushSize < minBrushSize {
		a.brushSize = minBrushSize
	}
}

func (a *ScrawlApp) handleMouseButtonEvent(e *sdl.MouseButtonEvent) {
	if e.Button != sdl.BUTTON_LEFT {
		return
	}
	switch e.Type {
	case sdl.MOUSEBUTTONDOWN:
		a.prevMouseX, a.prevMouseY = e.X, e.Y
		a.drawing = true

		// Accounts for the case where user only ks the mouse.
		// In such cases the line drawing algorithm might not proc
		// as no mouse motion is detected.
		a.drawCircle(a.prevMouseX, a.prevMouseY, a.brushSize)
	case sdl.MOUSEBUTTONUP:
		a.drawing = false
	}
}

func (a *ScrawlApp) handleMouseMotionEvent(e *sdl.MouseMotionEvent) {
	a.mouseX, a.mouseY = e.X, e.Y
	if a.drawing {
		a.drawLine(a.prevMouseX, a.prevMouseY, a.mouseX, a.mouseY)
		a.prevMouseX, a.prevMouseY = e.X, e.Y
	}
}

func (a *ScrawlApp) updateScreen() {
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
