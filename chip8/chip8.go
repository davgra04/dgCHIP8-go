package chip8

import (
	"fmt"
	"math/rand"
	"time"
)

// CHIP8 represents a CHIP8 machine with it's own memory, stack, buffers, etc
type CHIP8 struct {
	Cfg       *Config     // CHIP8 configuration
	program   []uint8     // program memory
	Display   []uint8     // display memory
	stack     []uint8     // stack memory
	registers []uint16    // register memory
	done      <-chan bool // signals CHIP8 to stop execution
	Cycle     uint        // number of cycles executed
}

// NewCHIP8 creates new CHIP8 machine given configuration
func NewCHIP8(cfg *Config) (*CHIP8, chan<- bool) {
	done := make(chan bool)
	chip := CHIP8{
		Cfg:       cfg,
		program:   make([]uint8, cfg.SizeProgram),
		Display:   make([]uint8, cfg.SizeDisplay),
		stack:     make([]uint8, cfg.SizeStack),
		registers: make([]uint16, cfg.SizeStack),
		done:      done,
		Cycle:     0,
	}

	return &chip, done
}

// RandomizeDisplay fills the display memory with random values
func (c *CHIP8) RandomizeDisplay() {
	for i := range c.Display {
		c.Display[i] = uint8(rand.Int() % 256)
	}
}

// DrawBinaryCount draws the current cycle number to screen in binary
func (c *CHIP8) DrawBinaryCount() {
	var byteToDraw uint8
	count := c.Cycle

	for byteIdx := range c.Display {

		byteToDraw = 0 // initialize byte to draw

		// fill byte to draw
		for i := 0; i < 8; i++ {
			byteToDraw = byteToDraw >> 1
			if count%2 == 1 {
				byteToDraw |= 0x80
			}
			count /= 2
		}

		c.Display[byteIdx] = byteToDraw // write to display memory

		// stop filling display memory if full count value is drawn
		if count == 0 {
			break
		}
	}
}

// Run executes the fetch/decode/execute loop at the config's ClockFreq
func (c *CHIP8) Run() {

	clockPeriod := time.Nanosecond * time.Duration(1000000000.0/c.Cfg.ClockFreq)
	fmt.Printf("CHIP8 clockPeriod: %v\n", clockPeriod)
	clockTicker := time.NewTicker(clockPeriod)
	defer clockTicker.Stop()

	running := true
	c.RandomizeDisplay()

	for running {
		select {
		case <-c.done:
			running = false
			break
		case <-clockTicker.C:
			c.DrawBinaryCount()
			c.Cycle++
			break
		}
	}

	fmt.Println("CHIP8 halted")
}
