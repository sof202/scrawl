package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	width        int32 = 800
	height       int32 = 600
	bytesPerRow  int32 = 4 // len("RGBA") = 4
	stride             = int(width * bytesPerRow)
	minBrushSize       = 1
	maxBrushSize       = 50
)

var (
	drawing          = false
	brushSize  int32 = 10
	prevMouseX       = width / 2
	prevMouseY       = height / 2
	mouseX           = width / 2
	mouseY           = height / 2
	white            = color.RGBA{255, 255, 255, 255}

	// The idea here is that, as we don't care about colours (only black
	// strokes on a white background), our canvas is just a vector of
	// black/white (Boolean). We then can just convert the canvas to pixels on
	// each frame.
	canvas = make([]bool, width*height) // false -> white, true -> black
	pixels = make([]byte, width*height*bytesPerRow)
)

// Main flow:
// Parse args -> Setup window -> Event listen -> Update drawn texture -> Render
func main() {
	// Parse args
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: scrawl <out.png>")
		os.Exit(1)
	}
	outputPath := os.Args[1]

	// Setup SDL2 window
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

	updateScreen := func() {
		canvasToPixels()
		texture.Update(
			nil,
			unsafe.Pointer(unsafe.SliceData(pixels)),
			stride,
		)
		renderer.Copy(texture, nil, nil)

		// Must come after copying texture as the texture covers the entire
		// window.
		drawRing(renderer, mouseX, mouseY, brushSize)
		renderer.Present()
	}

	// Event listener
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				saveImage(outputPath)

			case *sdl.KeyboardEvent:
				if e.Type != sdl.KEYDOWN {
					continue
				}
				switch e.Keysym.Sym {
				case sdl.K_c: // clear
					canvas = make([]bool, width*height)
				case sdl.K_ESCAPE: // exit without saving
					running = false
				}

			case *sdl.MouseWheelEvent:
				brushSize += e.Y * 4
				if brushSize > maxBrushSize {
					brushSize = maxBrushSize
				}
				if brushSize < minBrushSize {
					brushSize = minBrushSize
				}

			case *sdl.MouseButtonEvent:
				if e.Button != sdl.BUTTON_LEFT {
					continue
				}
				switch e.Type {
				case sdl.MOUSEBUTTONDOWN:
					prevMouseX, prevMouseY = e.X, e.Y
					drawing = true
				case sdl.MOUSEBUTTONUP:
					drawing = false
				}

			case *sdl.MouseMotionEvent:
				mouseX, mouseY = e.X, e.Y
				if drawing {
					drawLine(prevMouseX, prevMouseY, mouseX, mouseY)
					prevMouseX, prevMouseY = e.X, e.Y
				}
			}
		}
		updateScreen()
		sdl.Delay(16)
	}
}

// Our canvas holds whether each pixel is black/drawn or white/not-drawn. This
// isn't useful to the renderer however. As such, we need to convert this
// information into RGBA information for each pixel on the screen/texture.
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

// Updates the canvas to have a circle centred at the cartesian coordinate
// (cx,cy) with the given radius.
func drawCircle(cx, cy, radius int32) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > radius*radius { // circle defn: `x^2 + y^2 <= r^2`
				continue
			}
			x, y := cx+dx, cy+dy

			if x < 0 || x >= width || y < 0 || y >= height { // OOB
				continue
			}
			canvas[x+y*width] = true // black
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
func drawLine(x0, y0, x1, y1 int32) {
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
		drawCircle(x0, y0, brushSize)
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
func saveImage(path string) {
	// Images drawn in scawl only use black (drawn) and white (not-drawn). As
	// such, the image can be purely grayscale. This reduces some complexity
	// and gives a slight speedup.
	img := image.NewGray(image.Rect(0, 0, int(width), int(height)))

	for i, drawn := range canvas {
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
func drawRing(renderer *sdl.Renderer, cx, cy, radius int32) {
	renderer.SetDrawColor(128, 128, 128, 255) // gray
	const circleSegments = 16
	var prevX, prevY int32
	for i := 0; i <= circleSegments; i++ {
		angle := (float64(i) * 2 * math.Pi) / circleSegments
		x := cx + int32(float64(radius)*math.Cos(angle))
		y := cy + int32(float64(radius)*math.Sin(angle))
		if i > 0 {
			renderer.DrawLine(prevX, prevY, x, y)
		}
		prevX, prevY = x, y
	}
}
