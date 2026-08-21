package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	// Canvas is to be square, I like this more aesthetically
	defaultWidth  int32 = 600
	defaultHeight int32 = 600
)

type CLI struct {
	outputPath string
	width      int32
	height     int32
}

func usage() {
	fmt.Print(
		"Usage: scrawl <out.png>\n",
		"  -w, --width <n> The width of the window and output image\n",
		"  -h, --height <n>  The height of the window and output image\n",
	)
}

func parseArgs() CLI {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		usage()
		os.Exit(0)
	}
	if os.Args[1] == "-v" || os.Args[1] == "--version" {
		fmt.Println("scrawl:", version)
		os.Exit(0)
	}

	cli := CLI{width: defaultWidth, height: defaultHeight}

	cli.outputPath = os.Args[1]
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "-w" || os.Args[i] == "--width" {
			i++
			width, err := strconv.Atoi(os.Args[i])
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"width must be coercible to integer (%s)\n",
					os.Args[i],
				)
				usage()
				os.Exit(1)
			}
			cli.width = int32(width)
		} else if os.Args[i] == "-h" || os.Args[i] == "--height" {
			i++
			height, err := strconv.Atoi(os.Args[i])
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"height must be coercible to integer (%s)\n",
					os.Args[i],
				)
				usage()
				os.Exit(1)
			}
			cli.height = int32(height)
		} else {
			fmt.Fprintf(os.Stderr, "Not a valid option: %s\n", os.Args[i])
			usage()
			os.Exit(1)
		}
	}
	return cli
}
