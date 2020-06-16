package sdlio

import (
	"github.com/davgra04/dgCHIP8-go/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

// type CHIP8WindowConfig struct {

// }

func DrawCHIP8Display(window *sdl.Window, chip *chip8.CHIP8) {
	// clear display
	// renderer, err := window.GetRenderer()
	// if err != nil {
	// 	panic(err)
	// }
	// renderer.SetDrawColor(0, 0, 0, 255)
	// renderer.Clear()

	// draw pixels
	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	var pixelSize int32 = 14

	// fmt.Println("Drawing")

	for byteIdx, byte := range chip.Display {
		// fmt.Printf("    byte %03d:  %v\n", byteIdx, byte)
		for i := 0; i < 8; i++ {
			if byte%2 == 1 {
				x := int32((byteIdx*8 + i) % chip.Cfg.ResolutionX)
				y := int32((byteIdx*8 + i) / chip.Cfg.ResolutionX)
				rect := sdl.Rect{X: x * pixelSize, Y: y * pixelSize, W: pixelSize, H: pixelSize}
				surface.FillRect(&rect, 0xff00c0d3)
			}
			byte /= 2
		}

	}

}

func DrawCount(window *sdl.Window, chip *chip8.CHIP8, count int) {

	// draw pixels
	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	var pixelSize int32 = 14

	for i := 0; i < len(chip.Display)*8; i++ {
		// for i := range int(chip.Display) * 8 {
		if count%2 == 1 {
			x := int32(i % chip.Cfg.ResolutionX)
			y := int32(i / chip.Cfg.ResolutionX)
			rect := sdl.Rect{X: x * pixelSize, Y: y * pixelSize, W: pixelSize, H: pixelSize}
			surface.FillRect(&rect, 0xff00c0d3)
		}

		count /= 2

	}

}
