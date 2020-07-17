package chip8

import (
	"fmt"
	"testing"
	"time"
)

var chipCfg *Config
var chip *CHIP8
var done chan<- bool

////////////////////////////////////////////////////////////////////////////////
// tests
////////////////////////////////////////////////////////////////////////////////

// 00E0 - CLS
// Clear the display.

func TestInstructionClearScreen(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	for i := range chip.Display {
		chip.Display[i] = 0xba
	}

	chip.WriteShort(0x200, 0x00e0)

	chip.StepEmulation()

	for i := range chip.Display {
		if chip.Display[i] != 0 {
			t.Errorf("chip.Display[0x%x] = 0x%x; want 0", i, chip.Display[i])
		}
	}
}

// 2nnn - CALL addr
// Call subroutine at nnn.

// 00EE - RET
// Return from a subroutine.

func TestInstructionCallReturnSubroutine(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.WriteShort(0x200, 0x2abc)
	chip.WriteShort(0xabc, 0x00ee)

	var tests = []struct {
		StackPtr    uint8
		StackValue0 uint16
		PC          uint16
	}{
		{0x1, 0x200, 0xabc},
		{0x0, 0x200, 0x200},
		{0x1, 0x200, 0xabc},
		{0x0, 0x200, 0x200},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.StackPtr != want.StackPtr {
			t.Errorf("test %d: chip.StackPtr = 0x%x; want 0x%x", i, chip.StackPtr, want.StackPtr)
		}

		if chip.Stack[0] != want.StackValue0 {
			t.Errorf("test %d: chip.Stack[0] = 0x%x; want 0x%x", i, chip.Stack[0], want.StackValue0)
		}

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}
	}
}

// 1nnn - JP addr
// Jump to location nnn.

func TestInstructionJump(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.WriteShort(0x200, 0x1abc)
	chip.WriteShort(0xabc, 0x1def)

	var tests = []struct {
		PC uint16
	}{
		{0xabc},
		{0xdef},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}
	}
}

// 3xkk - SE Vx, byte
// Skip next instruction if Vx = kk.

func TestInstructionSkipEqualByte(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x1] = 0xba
	chip.Reg[0xa] = 0xdc

	chip.WriteShort(0x200, 0x31ba) // SE V1, 0xba	(should skip)
	chip.WriteShort(0x202, 0x1aaa) // jump away
	chip.WriteShort(0x204, 0x3add) // SE Va, 0xdd	(should not skip)

	var tests = []struct {
		PC uint16
	}{
		{0x204},
		{0x206},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}
	}
}

// 4xkk - SNE Vx, byte
// Skip next instruction if Vx != kk.

func TestInstructionSkipNotEqualByte(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x1] = 0xba
	chip.Reg[0xa] = 0xdc

	chip.WriteShort(0x200, 0x41bb) // SNE V1, 0xbb	(should skip)
	chip.WriteShort(0x202, 0x1aaa) // jump away
	chip.WriteShort(0x204, 0x4adc) // SNE Va, 0xdc	(should not skip)

	var tests = []struct {
		PC uint16
	}{
		{0x204},
		{0x206},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}
	}
}

// 5xy0 - SE Vx, Vy
// Skip next instruction if Vx = Vy.

func TestInstructionSkipEqualReg(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x1] = 0xba
	chip.Reg[0xa] = 0xdc
	chip.Reg[0xc] = 0xdc

	chip.WriteShort(0x200, 0x5ac0) // SE Va, Vc	(should skip)
	chip.WriteShort(0x202, 0x1aaa) // jump away
	chip.WriteShort(0x204, 0x51a0) // SE V1, Va	(should not skip)

	var tests = []struct {
		PC uint16
	}{
		{0x204},
		{0x206},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}
	}
}

// 6xkk - LD Vx, byte
// Set Vx = kk.

func TestInstructionLoadByte(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.WriteShort(0x200, 0x6a44)

	chip.StepEmulation()

	if chip.Reg[0xa] != 0x44 {
		t.Errorf("chip.Reg[0xa] = 0x%x; want 0x44", chip.Reg[0xa])
	}
}

// 7xkk - ADD Vx, byte
// Set Vx = Vx + kk.

func TestInstructionAddByte(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0x10
	chip.Reg[0x1] = 0xff

	chip.WriteShort(0x200, 0x7001)
	chip.WriteShort(0x202, 0x7101)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
	}{
		{0x202, 0x0, 0x11},
		{0x204, 0x1, 0x00},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}
	}
}

// 8xy0 - LD Vx, Vy
// Set Vx = Vy.
// func (chip *CHIP8) instructionLoadReg(instruction uint16) {

func TestInstructionLoadReg(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x1] = 0xff

	chip.WriteShort(0x200, 0x8210)

	chip.StepEmulation()

	if chip.Reg[0x2] != chip.Reg[0x1] {
		t.Errorf("chip.Reg[0x2] = 0x%x; want 0x%x", chip.Reg[0x2], chip.Reg[0x1])
	}
}

// 8xy1 - OR Vx, Vy
// Set Vx = Vx OR Vy.

func TestInstructionOr(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0xaa
	chip.Reg[0x1] = 0x55

	chip.WriteShort(0x200, 0x8011)
	chip.WriteShort(0x202, 0x8231)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
	}{
		{0x202, 0x0, 0xff},
		{0x204, 0x2, 0x00},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}
	}
}

// 8xy2 - AND Vx, Vy
// Set Vx = Vx AND Vy.

func TestInstructionAnd(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0xaa
	chip.Reg[0x1] = 0x55
	chip.Reg[0x2] = 0xff
	chip.Reg[0x3] = 0xaa

	chip.WriteShort(0x200, 0x8012)
	chip.WriteShort(0x202, 0x8232)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
	}{
		{0x202, 0x0, 0x00},
		{0x204, 0x2, 0xaa},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}
	}
}

// 8xy3 - XOR Vx, Vy
// Set Vx = Vx XOR Vy.

func TestInstructionXor(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0x3c // 0011 1100
	chip.Reg[0x1] = 0x0f // 0000 1111

	chip.WriteShort(0x200, 0x8013)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
	}{
		{0x202, 0x0, 0x33},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}
	}
}

// 8xy4 - ADD Vx, Vy
// Set Vx = Vx + Vy, set VF = carry.

func TestInstructionAddReg(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0x00
	chip.Reg[0x1] = 0x01
	chip.Reg[0x2] = 0xff
	chip.Reg[0x3] = 0x04

	chip.WriteShort(0x200, 0x8014)
	chip.WriteShort(0x202, 0x8234)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
		carry  uint8
	}{
		{0x202, 0x0, 0x01, 0x0},
		{0x204, 0x2, 0x03, 0x1},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}

		if chip.Reg[0xf] != want.carry {
			t.Errorf("test %d: chip.Reg[0xf] = 0x%x; want 0x%x", i, chip.Reg[0xf], want.carry)
		}
	}
}

// 8xy5 - SUB Vx, Vy
// Set Vx = Vx - Vy, set VF = NOT borrow.

func TestInstructionSubReg(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0xff
	chip.Reg[0x1] = 0x0f
	chip.Reg[0x2] = 0x02
	chip.Reg[0x3] = 0x04

	chip.WriteShort(0x200, 0x8015)
	chip.WriteShort(0x202, 0x8235)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
		carry  uint8
	}{
		{0x202, 0x0, 0xf0, 0x1},
		{0x204, 0x2, 0xfe, 0x0},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}

		if chip.Reg[0xf] != want.carry {
			t.Errorf("test %d: chip.Reg[0xf] = 0x%x; want 0x%x", i, chip.Reg[0xf], want.carry)
		}
	}
}

// 8xy6 - SHR Vx
// Set Vx = Vx SHR 1.

func TestInstructionShiftRight(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0xf0
	chip.Reg[0x1] = 0x0f

	chip.WriteShort(0x200, 0x8006)
	chip.WriteShort(0x202, 0x8106)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
		carry  uint8
	}{
		{0x202, 0x0, 0x78, 0x0},
		{0x204, 0x1, 0x07, 0x1},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}

		if chip.Reg[0xf] != want.carry {
			t.Errorf("test %d: chip.Reg[0xf] = 0x%x; want 0x%x", i, chip.Reg[0xf], want.carry)
		}
	}
}

// 8xy7 - SUBN Vx, Vy
// Set Vx = Vy - Vx, set VF = NOT borrow.

func TestInstructionSubNReg(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0x0f
	chip.Reg[0x1] = 0xff
	chip.Reg[0x2] = 0x04
	chip.Reg[0x3] = 0x02

	chip.WriteShort(0x200, 0x8017)
	chip.WriteShort(0x202, 0x8237)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
		carry  uint8
	}{
		{0x202, 0x0, 0xf0, 0x1},
		{0x204, 0x2, 0xfe, 0x0},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}

		if chip.Reg[0xf] != want.carry {
			t.Errorf("test %d: chip.Reg[0xf] = 0x%x; want 0x%x", i, chip.Reg[0xf], want.carry)
		}
	}
}

// 8xyE - SHL Vx {, Vy}
// Set Vx = Vx SHL 1.

func TestInstructionShiftLeft(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0xf0
	chip.Reg[0x1] = 0x0f

	chip.WriteShort(0x200, 0x800e)
	chip.WriteShort(0x202, 0x810e)

	var tests = []struct {
		PC     uint16
		regIdx uint8
		regVal uint8
		carry  uint8
	}{
		{0x202, 0x0, 0xe0, 0x1},
		{0x204, 0x1, 0x1e, 0x0},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}

		if chip.Reg[want.regIdx] != want.regVal {
			t.Errorf("test %d: chip.Reg[0x%x] = 0x%x; want 0x%x", i, want.regIdx, chip.Reg[want.regIdx], want.regVal)
		}

		if chip.Reg[0xf] != want.carry {
			t.Errorf("test %d: chip.Reg[0xf] = 0x%x; want 0x%x", i, chip.Reg[0xf], want.carry)
		}
	}
}

// 9xy0 - SNE Vx, Vy
// Skip next instruction if Vx != Vy.
// func (chip *CHIP8) instructionSkipNotEqualReg(instruction uint16) {

func TestInstructionSkipNotEqualReg(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x1] = 0xba
	chip.Reg[0x2] = 0xdc
	chip.Reg[0x3] = 0xdc

	chip.WriteShort(0x200, 0x9120) // SNE V1, V2	(should skip)
	chip.WriteShort(0x202, 0x1aaa) // jump away
	chip.WriteShort(0x204, 0x9ad0) // SNE V2, V3	(should not skip)

	var tests = []struct {
		PC uint16
	}{
		{0x204},
		{0x206},
	}

	for i, want := range tests {
		chip.StepEmulation()

		if chip.PC != want.PC {
			t.Errorf("test %d: chip.PC = 0x%x; want 0x%x", i, chip.PC, want.PC)
		}
	}
}

// Annn - LD I, addr
// Set I = nnn.

func TestInstructionLoadRegI(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.WriteShort(0x200, 0xabcd)

	chip.StepEmulation()

	if chip.RegI != 0xbcd {
		t.Errorf("chip.RegI = 0x%x; want 0xbcd", chip.Reg[0xa])
	}
}

// Bnnn - JP V0, addr
// Jump to location nnn + V0.

func TestInstructionJumpReg(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x0] = 0xff

	chip.WriteShort(0x200, 0xb400)

	chip.StepEmulation()

	if chip.PC != 0x4ff {
		t.Errorf("chip.PC = 0x%x; want 0x4ff", chip.PC)
	}
}

// Cxkk - RND Vx, byte
// Set Vx = random byte AND kk.

// skipping for now, not sure how to check if a random byte was written

// Dxyn - DRW Vx, Vy, nibble
// Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.

func printDisplay(chip *CHIP8) {
	for i := range chip.Display {
		if (i*8)%chip.Cfg.ResolutionX == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%08b ", chip.Display[i])
	}
}

func TestInstructionDrawSprite(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	// sprite data
	chip.WriteByte(0x400, 0b11110000)
	chip.WriteByte(0x401, 0b11110000)

	chip.RegI = 0x400
	chip.Reg[0x0] = 0
	chip.Reg[0x1] = 1
	chip.Reg[0x2] = 59

	chip.WriteShort(0x200, 0xd002)
	chip.WriteShort(0x202, 0xd112)

	// draw this first:
	//     11110000
	//     11110000
	//     00000000
	// printDisplay(chip)

	chip.StepEmulation()

	// fmt.Println(("\n\nAFTER 1ST DRAW"))
	// printDisplay(chip)
	// time.Sleep(time.Second)

	for i, pattern := range []uint8{
		0b11110000,
		0b11110000,
		0b00000000,
	} {
		dispIdx := i * chip.Cfg.ResolutionX / 8
		if chip.Display[dispIdx] != pattern {
			t.Errorf("chip.Display[%d] = 0b%x; want 0b%08b", dispIdx, chip.Display[dispIdx], pattern)
		}
	}

	// then draw this
	//     00000000
	//     01111000
	//     01111000

	// which should result in this
	//     11110000
	//     10001000
	//     01111000

	chip.StepEmulation()

	// fmt.Println(("\n\nAFTER 2ND DRAW"))
	// printDisplay(chip)
	// time.Sleep(time.Second)

	for i, pattern := range []uint8{
		0b11110000,
		0b10001000,
		0b01111000,
	} {
		dispIdx := i * chip.Cfg.ResolutionX / 8
		if chip.Display[dispIdx] != pattern {
			t.Errorf("chip.Display[%d] = 0b%x; want 0b%08b", dispIdx, chip.Display[dispIdx], pattern)
		}
	}

	chip.clearDisplay()

	chip.WriteByte(0x400, 0b10001001)
	chip.WriteByte(0x401, 0b01000110)
	chip.WriteByte(0x402, 0b00100110)
	chip.WriteByte(0x403, 0b00011001)

	chip.WriteShort(0x204, 0xd004)

	chip.StepEmulation()

	fmt.Println(("\n\nANOTHA DRAW"))
	printDisplay(chip)
	time.Sleep(time.Second)

	// if chip.PC != 0x4ff {
	// 	t.Errorf("chip.PC = 0x%x; want 0x4ff", chip.PC)
	// }
}

// Ex9E - SKP Vx
// Skip next instruction if key with the value of Vx is pressed.
// func (chip *CHIP8) instructionSkipKey(instruction uint16) {

// ExA1 - SKNP Vx
// Skip next instruction if key with the value of Vx is not pressed.
// func (chip *CHIP8) instructionSkipNotKey(instruction uint16) {

// Fx07 - LD Vx, DT
// Set Vx = delay timer value.

func TestInstructionReadDelayTimer(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.RegDelay = 0xba
	chip.WriteShort(0x200, 0xf107)

	chip.StepEmulation()

	if chip.Reg[0x1] != 0xba {
		t.Errorf("chip.Reg[0x1] = 0x%x; want 0xba", chip.Reg[0x1])
	}
}

// Fx0A - LD Vx, K
// Wait for a key press, store the value of the key in Vx.
// func (chip *CHIP8) instructionWaitForKey(instruction uint16) {

// Fx15 - LD DT, Vx
// Set delay timer = Vx.
// func (chip *CHIP8) instructionSetDelayTimer(instruction uint16) {

func TestInstructionSetDelayTimer(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x1] = 0xba

	chip.WriteShort(0x200, 0xf115)

	chip.StepEmulation()

	if chip.RegDelay != 0xba {
		t.Errorf("chip.RegDelay = 0x%x; want 0xba", chip.RegDelay)
	}
}

// Fx18 - LD ST, Vx
// Set sound timer = Vx.
// func (chip *CHIP8) instructionSetSoundTimer(instruction uint16) {

func TestInstructionSetSoundTimer(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Reg[0x1] = 0xba

	chip.WriteShort(0x200, 0xf118)

	chip.StepEmulation()

	if chip.RegSound != 0xba {
		t.Errorf("chip.RegSound = 0x%x; want 0xba", chip.RegSound)
	}
}

// Fx1E - ADD I, Vx
// Set I = I + Vx.
// func (chip *CHIP8) instructionAddRegI(instruction uint16) {

// Fx29 - LD F, Vx
// Set I = location of sprite for digit Vx.
// func (chip *CHIP8) instructionLoadSprite(instruction uint16) {

// func TestInstructionLoadRegI(t *testing.T) {
// 	chipCfg := GetDefaultConfig()
// 	chip, _ := NewCHIP8(chipCfg)

// 	chip.WriteShort(0x200, 0xabcd)

// 	chip.StepEmulation()

// 	if chip.RegI != 0xbcd {
// 		t.Errorf("chip.RegI = 0x%x; want 0xbcd", chip.Reg[0xa])
// 	}
// }

// Fx33 - LD B, Vx
// Store BCD representation of Vx in memory locations I, I+1, and I+2.
// func (chip *CHIP8) instructionLoadBCD(instruction uint16) {

// Fx55 - LD [I], Vx
// Store registers V0 through Vx in memory starting at location I.
// func (chip *CHIP8) instructionLoadMulti(instruction uint16) {

// Fx65 - LD Vx, [I]
// Read registers V0 through Vx from memory starting at location I.
// func (chip *CHIP8) instructionReadMulti(instruction uint16) {
