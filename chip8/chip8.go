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
	Keys         []bool // key state
	KeysPrev     []bool // previous key state
	watchingKeys bool   // used for Fx0A - LD Vx, K instruction
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
		Memory:       make([]uint8, cfg.SizeMemory),
		PC:           programStartAddr,
		MAR:          programStartAddr,
		Display:      make([]uint8, cfg.SizeDisplay),
		Stack:        make([]uint16, cfg.SizeStack),
		StackPtr:     0,
		Reg:          make([]uint8, cfg.SizeStack),
		RegI:         0,
		RegDelay:     0,
		RegSound:     0,
		Keys:         make([]bool, numKeys),
		KeysPrev:     make([]bool, numKeys),
		watchingKeys: false,
		Cycle:        0,
		Cfg:          cfg,
		done:         done,
		Paused:       false,
	}

	chip.reset()

	return &chip, done
}

func (chip *CHIP8) writeSpriteData() {
	// 0
	chip.WriteByte(0x00, 0b11110000)
	chip.WriteByte(0x01, 0b10010000)
	chip.WriteByte(0x02, 0b10010000)
	chip.WriteByte(0x03, 0b10010000)
	chip.WriteByte(0x04, 0b11110000)

	// 1
	chip.WriteByte(0x05, 0b00100000)
	chip.WriteByte(0x06, 0b01100000)
	chip.WriteByte(0x07, 0b00100000)
	chip.WriteByte(0x08, 0b00100000)
	chip.WriteByte(0x09, 0b01110000)

	// 2
	chip.WriteByte(0x0a, 0b11110000)
	chip.WriteByte(0x0b, 0b00010000)
	chip.WriteByte(0x0c, 0b11110000)
	chip.WriteByte(0x0d, 0b10000000)
	chip.WriteByte(0x0e, 0b11110000)

	// 3
	chip.WriteByte(0x0f, 0b11110000)
	chip.WriteByte(0x10, 0b00010000)
	chip.WriteByte(0x11, 0b11110000)
	chip.WriteByte(0x12, 0b00010000)
	chip.WriteByte(0x13, 0b11110000)

	// 4
	chip.WriteByte(0x14, 0b10010000)
	chip.WriteByte(0x15, 0b10010000)
	chip.WriteByte(0x16, 0b11110000)
	chip.WriteByte(0x17, 0b00010000)
	chip.WriteByte(0x18, 0b00010000)

	// 5
	chip.WriteByte(0x19, 0b11110000)
	chip.WriteByte(0x1a, 0b10000000)
	chip.WriteByte(0x1b, 0b11110000)
	chip.WriteByte(0x1c, 0b00010000)
	chip.WriteByte(0x1d, 0b11110000)

	// 6
	chip.WriteByte(0x1e, 0b11110000)
	chip.WriteByte(0x1f, 0b10000000)
	chip.WriteByte(0x20, 0b11110000)
	chip.WriteByte(0x21, 0b10010000)
	chip.WriteByte(0x22, 0b11110000)

	// 7
	chip.WriteByte(0x23, 0b11110000)
	chip.WriteByte(0x24, 0b00010000)
	chip.WriteByte(0x25, 0b00100000)
	chip.WriteByte(0x26, 0b01000000)
	chip.WriteByte(0x27, 0b01000000)

	// 8
	chip.WriteByte(0x28, 0b11110000)
	chip.WriteByte(0x29, 0b10010000)
	chip.WriteByte(0x2a, 0b11110000)
	chip.WriteByte(0x2b, 0b10010000)
	chip.WriteByte(0x2c, 0b11110000)

	// 9
	chip.WriteByte(0x2d, 0b11110000)
	chip.WriteByte(0x2e, 0b10010000)
	chip.WriteByte(0x2f, 0b11110000)
	chip.WriteByte(0x30, 0b00010000)
	chip.WriteByte(0x31, 0b11110000)

	// a
	chip.WriteByte(0x32, 0b11110000)
	chip.WriteByte(0x33, 0b10010000)
	chip.WriteByte(0x34, 0b11110000)
	chip.WriteByte(0x35, 0b10010000)
	chip.WriteByte(0x36, 0b10010000)

	// b
	chip.WriteByte(0x37, 0b11110000)
	chip.WriteByte(0x38, 0b10010000)
	chip.WriteByte(0x39, 0b11100000)
	chip.WriteByte(0x3a, 0b10010000)
	chip.WriteByte(0x3b, 0b11110000)

	// c
	chip.WriteByte(0x3c, 0b11110000)
	chip.WriteByte(0x3d, 0b10000000)
	chip.WriteByte(0x3e, 0b10000000)
	chip.WriteByte(0x3f, 0b10000000)
	chip.WriteByte(0x40, 0b11110000)

	// d
	chip.WriteByte(0x41, 0b11100000)
	chip.WriteByte(0x42, 0b10010000)
	chip.WriteByte(0x43, 0b10010000)
	chip.WriteByte(0x44, 0b10010000)
	chip.WriteByte(0x45, 0b11100000)

	// e
	chip.WriteByte(0x46, 0b11110000)
	chip.WriteByte(0x47, 0b10000000)
	chip.WriteByte(0x48, 0b11110000)
	chip.WriteByte(0x49, 0b10000000)
	chip.WriteByte(0x4a, 0b11110000)

	// f
	chip.WriteByte(0x4b, 0b11110000)
	chip.WriteByte(0x4c, 0b10000000)
	chip.WriteByte(0x4d, 0b11110000)
	chip.WriteByte(0x4e, 0b10000000)
	chip.WriteByte(0x4f, 0b10000000)
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
	chip.writeSpriteData()
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
	if addr < chip.Cfg.SizeDisplay {
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
// key state functions
////////////////////////////////////////////////////////////////////////////////

// SetKeyState updates the state of a key and previous key state array
func (chip *CHIP8) SetKeyState(key uint8, state bool) {
	if key >= 16 {
		return
	}

	// set current key state
	chip.Keys[key] = state
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
	// decrement timers at 60Hz
	if chip.Cycle%60 == 0 {
		if chip.RegDelay > 0 {
			chip.RegDelay--
		}
		if chip.RegSound > 0 {
			chip.RegSound--
		}
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
