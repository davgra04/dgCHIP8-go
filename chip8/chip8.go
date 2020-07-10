package chip8

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	programStartAddr = 0x200
	numKeys          = 16
)

// CHIP8 represents a CHIP8 machine with it's own memory, stack, buffers, etc
type CHIP8 struct {
	// memory
	Memory []uint8 // program memory
	PC     uint16  // program counter
	// display
	Display []uint8 // display memory
	// stack
	Stack    []uint16 // stack memory
	StackPtr uint8    // pointer to head of the stack
	// registers
	Registers []uint8 // register memory
	RegI      uint8   // I register
	RegDelay  uint8   // delay register
	RegSound  uint8   // sound register
	// keys
	Keys []bool // key state
	// etc
	Cycle uint64      // number of cycles executed
	Cfg   *Config     // CHIP8 configuration
	done  <-chan bool // signals CHIP8 to stop execution
}

// NewCHIP8 creates new CHIP8 machine given configuration
func NewCHIP8(cfg *Config) (*CHIP8, chan<- bool) {
	done := make(chan bool)
	chip := CHIP8{
		Memory:    make([]uint8, cfg.SizeMemory),
		PC:        0x200,
		Display:   make([]uint8, cfg.SizeDisplay),
		Stack:     make([]uint16, cfg.SizeStack),
		StackPtr:  0,
		Registers: make([]uint8, cfg.SizeStack),
		RegI:      0,
		RegDelay:  0,
		RegSound:  0,
		Keys:      make([]bool, numKeys),
		Cycle:     0,
		Cfg:       cfg,
		done:      done,
	}

	return &chip, done
}

func (chip *CHIP8) clearMemory() {
	for i := range chip.Memory {
		chip.Memory[i] = 0x00
	}
}

func (chip *CHIP8) clearRegisters() {
	for i := range chip.Registers {
		chip.Registers[i] = 0
	}
	chip.RegI = 0
	chip.RegDelay = 0
	chip.RegSound = 0
}

func (chip *CHIP8) clearStack() {
	for i := range chip.Stack {
		chip.Stack[i] = 0
	}
	chip.StackPtr = 0
}

func (chip *CHIP8) clearDisplay() {
	for i := range chip.Display {
		chip.Display[i] = 0
	}
}

func (chip *CHIP8) reset() {
	chip.clearMemory()
	chip.clearDisplay()
	chip.clearRegisters()
	chip.clearStack()
	chip.PC = 0x200
	chip.Cycle = 0

}

// LoadProgram initializes the CHIP8's memory with the program
func (chip *CHIP8) LoadProgram(program []byte) {
	chip.reset()
	for i := range program {
		chip.Memory[i+programStartAddr] = program[i]
	}
}

// ReadWord returns a 2 byte word from the specified address
func (chip *CHIP8) ReadWord(addr uint16) uint16 {
	return uint16(chip.Memory[addr])<<8 + uint16(chip.Memory[addr+1])
}

////////////////////////////////////////////////////////////////////////////////
// test draw functions
////////////////////////////////////////////////////////////////////////////////

// RandomizeDisplay fills the display memory with random values
func (chip *CHIP8) RandomizeDisplay() {
	for i := range chip.Display {
		chip.Display[i] = uint8(rand.Int() % 256)
	}
}

// DrawBinaryCount draws the current cycle number to screen in binary
func (chip *CHIP8) DrawBinaryCount() {
	var byteToDraw uint8
	count := chip.Cycle

	for byteIdx := range chip.Display {

		byteToDraw = 0 // initialize byte to draw

		// fill byte to draw
		for i := 0; i < 8; i++ {
			byteToDraw = byteToDraw >> 1
			if count%2 == 1 {
				byteToDraw |= 0x80
			}
			count /= 2
		}

		chip.Display[byteIdx] = byteToDraw // write to display memory

		// stop filling display memory if full count value is drawn
		if count == 0 {
			break
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// main execution loop
////////////////////////////////////////////////////////////////////////////////

// Run executes the fetch/decode/execute loop at the config's ClockFreq
func (chip *CHIP8) Run() {

	clockPeriod := time.Nanosecond * time.Duration(1000000000.0/chip.Cfg.ClockFreq)
	fmt.Printf("CHIP8 clockPeriod: %v\n", clockPeriod)
	clockTicker := time.NewTicker(clockPeriod)
	defer clockTicker.Stop()

	running := true
	chip.RandomizeDisplay()

	for running {
		select {
		case <-chip.done:
			running = false
			break
		case <-clockTicker.C:
			chip.DrawBinaryCount()
			chip.Cycle++
			break
		}
	}

	fmt.Println("CHIP8 halted")
}
