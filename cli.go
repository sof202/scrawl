package main

import (
	"fmt"
	"strconv"
)

const (
	// Canvas is to be square, I like this more aesthetically
	defaultWidth  int32 = 600
	defaultHeight int32 = 600

	noExitCode = -1 // Sentinel value
)

var (
	// For releases, this should be overriden with:
	//   `-ldflags "-X main.version=$(git describe --tags)"`
	version = "dev"
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

func parseArgs(args []string) (cli CLI, exitCode int, err error) {
	if len(args) < 2 {
		return CLI{}, 1, fmt.Errorf("missing required <out.png> argument")
	}

	switch args[1] {
	case "-h", "--help":
		usage()
		return CLI{}, 0, nil
	case "-v", "--version":
		fmt.Println("scrawl:", version)
		return CLI{}, 0, nil
	}

	cli = CLI{
		outputPath: args[1],
		width:      defaultWidth,
		height:     defaultHeight,
	}

	i := 2
	nextArg := func() (string, bool) {
		i++
		if i >= len(args) {
			return "", false
		}
		return args[i], true
	}
	for ; i < len(args); i++ {
		switch args[i] {
		case "-w", "--width":
			value, ok := nextArg()
			if !ok {
				return CLI{}, 1, fmt.Errorf("No argument for %s", args[i-1])
			}
			width, err := strconv.Atoi(value)
			if err != nil {
				return CLI{}, 1, fmt.Errorf("width must be coercible to integer (%s)", value)
			}
			cli.width = int32(width)
		case "-h", "--height":
			value, ok := nextArg()
			if !ok {
				return CLI{}, 1, fmt.Errorf("No argument for %s", args[i-1])
			}
			height, err := strconv.Atoi(value)
			if err != nil {
				return CLI{}, 1, fmt.Errorf("height must be coercible to integer (%s)", value)
			}
			cli.width = int32(height)
		default:
			return CLI{}, 1, fmt.Errorf("Not a valid option %s", args[i])
		}
	}
	return cli, noExitCode, nil
}
