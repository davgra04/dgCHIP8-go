package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/davgra04/dgCHIP8-go/chip8"
	"github.com/davgra04/dgCHIP8-go/sdlio"
	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	// Relevant issue: https://github.com/golang/go/issues/23112
	runtime.LockOSThread()
}

func main() {
	// initialize SDL
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// create window
	window, err := sdl.CreateWindow("dgchip8", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		896, 448, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	// set up CHIP8 machine
	chipCfg := chip8.GetDefaultConfig()
	// chipCfg.ClockFreq = 2000.0
	chip, done := chip8.NewCHIP8(chipCfg)

	// prepare display refresh timer
	displayPeriod := time.Microsecond * time.Duration(1000000.0/chipCfg.ScreenRefreshFreq)
	fmt.Printf("displayPeriod: %v\n", displayPeriod)
	displayTicker := time.NewTicker(displayPeriod)
	defer displayTicker.Stop()

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

		// draw CHIP8 display
		sdlio.DrawCHIP8Display(window, chip)
		window.UpdateSurface()

		// wait until next draw
		<-displayTicker.C
	}
}
