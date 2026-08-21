package main

import (
	"fmt"
	"log"
	"os"
)

const (
	bytesPerRow int32 = 4 // len("RGBA") = 4

	minBrushSize int32 = 1
	maxBrushSize int32 = 50
)

// Main flow:
// Parse args -> Setup window -> Event listen -> Update drawn texture -> Render
func main() {
	cli, exitCode, err := parseArgs(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		usage()
		os.Exit(exitCode)
	}
	if exitCode >= 0 {
		os.Exit(exitCode)
	}

	app, err := NewApp(cli)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
