package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	bytesPerRow   int32 = 4 // len("RGBA") = 4
	defaultWidth  int32 = 800
	defaultHeight int32 = 600
	minBrushSize  int32 = 1
	maxBrushSize  int32 = 50
)

// Main flow:
// Parse args -> Setup window -> Event listen -> Update drawn texture -> Render
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: scrawl <out.png>")
		os.Exit(1)
	}

	app, err := newApp(defaultWidth, defaultHeight)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	if err := app.Run(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

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

func newApp(width, height int32) (*ScrawlApp, error) {
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

// Our canvas holds whether each pixel is black/drawn or white/not-drawn. This
// isn't useful to the renderer however. As such, we need to convert this
// information into RGBA information for each pixel on the screen/texture.
func (a *ScrawlApp) canvasToPixels() {
	for i, drawn := range a.canvas {
		idx := i * int(bytesPerRow)
		var v byte
		if !drawn {
			v = 255 // white
		}
		a.pixels[idx+0] = v
		a.pixels[idx+1] = v
		a.pixels[idx+2] = v
		a.pixels[idx+3] = 255 // always no alpha/transparency
	}

}

// Updates the canvas to have a circle centred at the cartesian coordinate
// (cx,cy) with the given radius.
func (a *ScrawlApp) drawCircle(cx, cy, radius int32) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > radius*radius { // circle defn: `x^2 + y^2 <= r^2`
				continue
			}
			x, y := cx+dx, cy+dy

			if x < 0 || x >= a.width || y < 0 || y >= a.height { // OOB
				continue
			}
			a.canvas[x+y*a.width] = true // black
		}
	}
}

// from psuedocode in:
//
//	https://en.wikipedia.org/wiki/Bresenham's_line_algorithm#All_cases
//
// This was required as drawCircle on it's own (being called continuously
// whilst mouse is held) results in skipping (due to the speed of processing).
// Rather than optimising everything, it's easier to use a line drawing
// algorithm such as this one (which is good enough as anti-aliasing isn't
// desirable here).
func (a *ScrawlApp) drawLine(x0, y0, x1, y1 int32) {
	var (
		dx, dy, sx, sy, error, errorDoubled int32
	)

	abs := func(x int32) int32 {
		if x < 0 {
			return -x
		}
		return x
	}

	dx = abs(x1 - x0)
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}

	dy = -abs(y1 - y0)
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}

	error = dx + dy

	for {
		a.drawCircle(x0, y0, a.brushSize)
		errorDoubled = 2 * error
		if errorDoubled >= dy {
			if x0 == x1 {
				break
			}
			error += dy
			x0 += sx
		}
		if errorDoubled <= dx {
			if y0 == y1 {
				break
			}
			error += dx
			y0 += sy
		}
	}
}

// Opens the given file path and encodes the current state of the canvas as a
// png. Images are generally between 1KB-20KB depending on how varied the image
// is (for more information look into the PNG file format).
func (a *ScrawlApp) saveImage(path string) {
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
		fmt.Fprintln(os.Stderr, "Saving failed:", err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		fmt.Fprintln(os.Stderr, "Encoding failed:", err)
	}
}

// Draws radial line segments around the position given by cartesian
// coordinates (cx,cy) with the given radius. The primary purpose being:
// Drawing a circle around the cursor indicating where 'paint' will be placed
// on the screen.
func (a *ScrawlApp) drawRing(cx, cy, radius int32) {
	a.renderer.SetDrawColor(128, 128, 128, 255) // gray
	const circleSegments = 16
	var prevX, prevY int32
	for i := 0; i <= circleSegments; i++ {
		angle := (float64(i) * 2 * math.Pi) / circleSegments
		x := cx + int32(float64(radius)*math.Cos(angle))
		y := cy + int32(float64(radius)*math.Sin(angle))
		if i > 0 {
			a.renderer.DrawLine(prevX, prevY, x, y)
		}
		prevX, prevY = x, y
	}
}
