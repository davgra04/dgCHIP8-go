package main

import (
	"fmt"
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

func main() {

	// Initialize
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

	// fontPath := "SourceCodePro-Regular.ttf"
	// fontPath := "SourceCodePro-Medium.ttf"
	// font, err := ttf.OpenFont(fontPath, 16)
	// if err != nil {
	// 	panic(err)
	// }

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

	// set up CHIP8 machine
	chipCfg := chip8.GetDefaultConfig()
	// chipCfg.ClockFreq = 2000.0
	chip, done := chip8.NewCHIP8(chipCfg)

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
			switch event.(type) {
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
