package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/davgra04/dgCHIP8-go/chip8"
	"github.com/davgra04/dgCHIP8-go/sdlio"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	// Relevant issue: https://github.com/golang/go/issues/23112
	runtime.LockOSThread()
}

func readProgram(path string) []byte {
	// file, err := os.Open(path)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	program, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("program: %v [%T]\n", program, program)
	return program
}

func main() {
	// Process command line arguments
	////////////////////////////////////////

	// usage: bin2go.py [-h] [-w WIDTH] infile

	// positional arguments:
	//   infile                file to convert into a Go bytes array

	// optional arguments:
	//   -h, --help            show this help message and exit
	//   -w WIDTH, --width WIDTH
	// 						sets maximum width in characters of output Go bytes
	// 						array

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-h] [-start_paused] program_path\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "positional arguments:\n")
		fmt.Fprintf(os.Stderr, "  program_path\n")
		fmt.Fprintf(os.Stderr, "        path of the program to load and execute:\n")
		fmt.Fprintf(os.Stderr, "optional arguments:\n")
		flag.PrintDefaults()
	}

	startPaused := flag.Bool("start_paused", false, "if set, CHIP8 machine will start paused")
	flag.Parse()

	if flag.NArg() < 1 {
		panic(fmt.Errorf("Must provide program_path to load into CHIP8"))
	}

	fmt.Printf("startPaused: %v [%T]\n", *startPaused, *startPaused)
	fmt.Printf("flag.Arg(0): %v [%T]\n", flag.Arg(0), flag.Arg(0))

	// Initialize CHIP8 machine
	////////////////////////////////////////

	chipCfg := chip8.GetDefaultConfig()
	// chipCfg.ClockFreq = 2000.0
	chip, done := chip8.NewCHIP8(chipCfg)

	// read program and load into CHIP8 memory
	chip.LoadProgram(readProgram(flag.Arg(0)))

	// Initialize SDL and CHIP8 machine
	////////////////////////////////////////

	// get default SDL window config
	windowCfg := sdlio.GetDefaultWindowConfig()

	// initialize SDL
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// initialize SDL ttf
	if err := ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()

	// load font
	font, err := sdlio.LoadFont(16)
	if err != nil {
		panic(err)
	}
	defer font.Close()

	// create window
	window, err := sdl.CreateWindow("dgchip8", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		windowCfg.Width, windowCfg.Height, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	// create renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	// wrap everything in SDLAppContext
	ctx := sdlio.SDLAppContext{
		WinCfg:   windowCfg,
		Window:   window,
		Renderer: renderer,
		Chip8:    chip,
		Font:     font,
	}

	// prepare display refresh timer
	displayPeriod := time.Microsecond * time.Duration(1000000.0/windowCfg.RefreshRate)
	fmt.Printf("displayPeriod: %v\n", displayPeriod)
	displayTicker := time.NewTicker(displayPeriod)
	defer displayTicker.Stop()

	// Run
	////////////////////////////////////////

	// start CHIP8 machine
	go chip.Run()

	// main loop
	running := true
	for running {
		// handle SDL events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				sdlio.HandleKey(&ctx, t)
				break
			case *sdl.QuitEvent:
				fmt.Println("Quit")
				running = false
				done <- true
				break
				// default:
				// 	fmt.Printf("event %s\n", event)
			}
		}

		// draw window
		sdlio.DrawWindow(&ctx)

		// wait until next draw
		<-displayTicker.C
	}
}
