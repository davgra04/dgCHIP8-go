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

	// create CHIP8 machine
	chipConfig := chip8.GetDefaultConfig()
	chip := chip8.NewCHIP8(chipConfig)

	// // draw a red square
	// surface, err := window.GetSurface()
	// if err != nil {
	// 	panic(err)
	// }
	// surface.FillRect(nil, 0)

	// rect := sdl.Rect{X: 300, Y: 200, W: 200, H: 200}
	// surface.FillRect(&rect, 0xff00c0d3)
	// window.UpdateSurface()

	displayTicker := time.NewTicker(time.Millisecond * time.Duration(1000.0/(chipConfig.ClockFreq)))
	// displayTicker := time.NewTicker(time.Second)

	// main loop
	running := true
	for running {
		// handle SDL events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			default:
				fmt.Printf("event %s\n", event)
			}
		}

		// randomize chip8 display
		chip.RandomizeDisplay()

		// draw chip8 display
		sdlio.DrawCHIP8Display(window, chip)
		window.UpdateSurface()

		// sleep until
		// time.Sleep(time.Millisecond * 16)
		<-displayTicker.C
	}
}
