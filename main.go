package main

import "log"

var (
	// For releases, this should be overriden with:
	//   `-ldflags "-X main.version=$(git describe --tags)"`
	version = "dev"
)

const (
	bytesPerRow int32 = 4 // len("RGBA") = 4

	minBrushSize int32 = 1
	maxBrushSize int32 = 50
)

// Main flow:
// Parse args -> Setup window -> Event listen -> Update drawn texture -> Render
func main() {
	app, err := NewApp(parseArgs())
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
