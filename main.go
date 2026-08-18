package main

import (
	"fmt"
	"log"
	"os"
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

	app, err := NewApp(os.Args[1], defaultWidth, defaultHeight)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
