package main

import (
	"fmt"
	"log"
	"os"
)

const (
	bytesPerRow int32 = 4 // len("RGBA") = 4

	// Canvas is to be square, I like this more aesthetically
	defaultWidth  int32 = 600
	defaultHeight int32 = 600
	minBrushSize  int32 = 1
	maxBrushSize  int32 = 50
)

var (
	// For releases, this should be overriden with:
	//   `-ldflags "-X main.version=$(git describe --tags)"`
	version = "dev"
)

// Main flow:
// Parse args -> Setup window -> Event listen -> Update drawn texture -> Render
func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println("Usage: scrawl <out.png>")
		os.Exit(0)
	}

	if os.Args[1] == "-v" || os.Args[1] == "--version" {
		fmt.Println("scawl:", version)
		os.Exit(0)
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
