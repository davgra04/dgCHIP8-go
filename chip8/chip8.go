package chip8

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	programStartAddr = 0x200 // location of first instruction in memory
	numKeys          = 16
)

// CHIP8 represents a CHIP8 machine with it's own memory, stack, buffers, etc
type CHIP8 struct {
	// memory
	Memory []uint8 // program memory
	PC     uint16  // program counter
	MAR    uint16  // memory address register
	// display
	Display []uint8 // display memory
	// stack
	Stack    []uint16 // stack memory
	StackPtr uint8    // pointer to head of the stack
	// registers
	Reg      []uint8 // register memory
	RegI     uint16  // I register
	RegDelay uint8   // delay register
	RegSound uint8   // sound register
	// keys
	Keys []bool // key state
	// etc
	Cycle  uint64      // number of cycles executed
	Cfg    *Config     // CHIP8 configuration
	done   <-chan bool // signals CHIP8 to stop execution
	Paused bool        // if true, pauses execution
}

// NewCHIP8 creates new CHIP8 machine given configuration
func NewCHIP8(cfg *Config) (*CHIP8, chan<- bool) {
	done := make(chan bool)
	chip := CHIP8{
		Memory:   make([]uint8, cfg.SizeMemory),
		PC:       programStartAddr,
		MAR:      programStartAddr,
		Display:  make([]uint8, cfg.SizeDisplay),
		Stack:    make([]uint16, cfg.SizeStack),
		StackPtr: 0,
		Reg:      make([]uint8, cfg.SizeStack),
		RegI:     0,
		RegDelay: 0,
		RegSound: 0,
		Keys:     make([]bool, numKeys),
		Cycle:    0,
		Cfg:      cfg,
		done:     done,
		Paused:   false,
	}

	return &chip, done
}

func (chip *CHIP8) clearMemory() {
	for i := range chip.Memory {
		chip.Memory[i] = 0x00
	}
}

func (chip *CHIP8) clearRegisters() {
	for i := range chip.Reg {
		chip.Reg[i] = 0
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
	chip.PC = programStartAddr
	chip.Cycle = 0

}

// LoadProgram initializes the CHIP8's memory with the program
func (chip *CHIP8) LoadProgram(program []byte) {
	chip.reset()
	for i := range program {
		chip.Memory[i+programStartAddr] = program[i]
	}
}

////////////////////////////////////////////////////////////////////////////////
// memory read/write functions
////////////////////////////////////////////////////////////////////////////////

// ReadByte returns a byte from the specified address
func (chip *CHIP8) ReadByte(addr uint16) uint8 {
	if addr > chip.Cfg.SizeMemory-1 {
		return 0xff
	}
	return chip.Memory[addr]
}

// ReadShort returns a short (2 bytes) from the specified address
func (chip *CHIP8) ReadShort(addr uint16) uint16 {
	if addr > chip.Cfg.SizeMemory-2 {
		return 0xffff
	}
	return uint16(chip.Memory[addr])<<8 + uint16(chip.Memory[addr+1])
}

// WriteByte writes a byte to program memory at the specified address
func (chip *CHIP8) WriteByte(addr uint16, value uint8) {
	if addr < chip.Cfg.SizeMemory-1 {
		chip.Memory[addr] = value
	}
}

// WriteShort writes a short (2 bytes) to program memory at the specified address
func (chip *CHIP8) WriteShort(addr uint16, value uint16) {
	if addr < chip.Cfg.SizeMemory-2 {
		chip.Memory[addr] = uint8(value >> 8 & 0xff)
		chip.Memory[addr+1] = uint8(value & 0xff)
	}
}

////////////////////////////////////////////////////////////////////////////////
// display read/write functions
////////////////////////////////////////////////////////////////////////////////

// ReadDisplayByte returns a byte from the specified address
func (chip *CHIP8) ReadDisplayByte(addr uint16) uint8 {
	if addr > chip.Cfg.SizeDisplay-1 {
		return 0xff
	}
	return chip.Display[addr]
}

// ReadDisplayShort returns a short (2 bytes) from the specified address
func (chip *CHIP8) ReadDisplayShort(addr uint16) uint16 {
	if addr > chip.Cfg.SizeDisplay-2 {
		return 0xffff
	}
	return uint16(chip.Display[addr])<<8 + uint16(chip.Display[addr+1])
}

// WriteDisplayByte writes a byte to display memory at the specified address
func (chip *CHIP8) WriteDisplayByte(addr uint16, value uint8) {
	if addr < chip.Cfg.SizeDisplay-1 {
		chip.Display[addr] = value
	}
}

// WriteDisplayShort writes a short (2 bytes) to display memory at the specified address
func (chip *CHIP8) WriteDisplayShort(addr uint16, value uint16) {
	if addr < chip.Cfg.SizeDisplay-2 {
		chip.Display[addr] = uint8(value >> 8)
		chip.Display[addr+1] = uint8(value)
	}
}

////////////////////////////////////////////////////////////////////////////////
// stack functions
////////////////////////////////////////////////////////////////////////////////

func (chip *CHIP8) pushStack(value uint16) {
	if chip.StackPtr < chip.Cfg.SizeStack {
		chip.Stack[chip.StackPtr] = value
		chip.StackPtr++
	}
}

func (chip *CHIP8) popStack() uint16 {
	if chip.StackPtr > 0 {
		chip.StackPtr--
		return chip.Stack[chip.StackPtr]
	}
	return 0xff
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

// StepEmulation executes a single fetch-decode-execute cycle
func (chip *CHIP8) StepEmulation() {
	// decrement timers
	if chip.RegDelay > 0 {
		chip.RegDelay--
	}
	if chip.RegSound > 0 {
		chip.RegSound--
	}

	// fetch and increment program counter
	chip.MAR = chip.PC
	chip.PC += 2

	// decode and execute
	instruction := chip.ReadShort(chip.MAR)
	chip.decodeAndExecuteInstruction(instruction)
	chip.Cycle++
}

// Run executes the fetch/decode/execute loop at the config's ClockFreq
func (chip *CHIP8) Run() {

	clockPeriod := time.Nanosecond * time.Duration(1000000000.0/chip.Cfg.ClockFreq)
	fmt.Printf("CHIP8 clockPeriod: %v\n", clockPeriod)
	clockTicker := time.NewTicker(clockPeriod)
	defer clockTicker.Stop()

	running := true
	// chip.RandomizeDisplay()

	for running {
		select {
		case <-chip.done:
			running = false
			break
		case <-clockTicker.C:
			if !chip.Paused {
				chip.StepEmulation()
			}
			break
		}
	}

	fmt.Println("CHIP8 halted")
}
