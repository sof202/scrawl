package main

import "math"

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

	dx = int32Abs(x1 - x0)
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}

	dy = -int32Abs(y1 - y0)
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

func int32Abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
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
