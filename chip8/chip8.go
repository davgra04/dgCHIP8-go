package chip8

import "math/rand"

// CHIP8 represents a CHIP8 machine with it's own memory, stack, buffers, etc
type CHIP8 struct {
	Cfg       *Config
	program   []uint8
	Display   []uint8
	stack     []uint8
	registers []uint16
}

// NewCHIP8 creates new CHIP8 machine given configuration
func NewCHIP8(cfg *Config) *CHIP8 {
	return &CHIP8{
		Cfg:       cfg,
		program:   make([]uint8, cfg.SizeProgram),
		Display:   make([]uint8, cfg.SizeDisplay),
		stack:     make([]uint8, cfg.SizeStack),
		registers: make([]uint16, cfg.SizeStack),
	}
}

// RandomizeDisplay fills the display memory with random values
func (c *CHIP8) RandomizeDisplay() {
	for i := range c.Display {
		c.Display[i] = uint8(rand.Int() % 256)
	}
}
