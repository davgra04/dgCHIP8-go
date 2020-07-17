package chip8

import (
	"testing"
)

////////////////////////////////////////////////////////////////////////////////
// tests
////////////////////////////////////////////////////////////////////////////////

func TestReset(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.PC = 42
	chip.Cycle = 42
	chip.Memory[0] = 42
	chip.Memory[42] = 42
	chip.Memory[chipCfg.SizeMemory-1] = 42
	chip.Display[0] = 42
	chip.Display[42] = 42
	chip.Display[chipCfg.SizeDisplay-1] = 42
	chip.Stack[0] = 42
	chip.Stack[5] = 42
	chip.Stack[chipCfg.SizeStack-1] = 42
	chip.StackPtr = chipCfg.SizeStack - 1
	chip.Reg[0] = 42
	chip.Reg[5] = 42
	chip.Reg[chipCfg.NumRegisters-1] = 42
	chip.RegI = 42
	chip.RegSound = 42
	chip.RegDelay = 42

	chip.reset()

	if chip.PC != programStartAddr {
		t.Errorf("chip.PC = 0x%x; want 0x%x", chip.PC, programStartAddr)
	}

	if chip.Cycle != 0 {
		t.Errorf("chip.Cycle = %d; want 0", chip.Cycle)
	}

	for i := range chip.Memory {
		if chip.Memory[i] != 0 {
			t.Errorf("chip.Memory[0x%x] = 0x%x; want 0x0", i, chip.Memory[i])
		}
	}

	for i := range chip.Display {
		if chip.Display[i] != 0 {
			t.Errorf("chip.Display[0x%x] = 0x%x; want 0x0", i, chip.Display[i])
		}
	}

	for i := range chip.Stack {
		if chip.Stack[i] != 0 {
			t.Errorf("chip.Stack[0x%x] = 0x%x; want 0x0", i, chip.Stack[i])
		}
	}

	if chip.StackPtr != 0 {
		t.Errorf("chip.StackPtr = 0x%x; want 0x0", chip.StackPtr)
	}

	for i := range chip.Reg {
		if chip.Reg[i] != 0 {
			t.Errorf("chip.Reg[0x%x] = 0x%x; want 0x0", i, chip.Reg[i])
		}
	}

	if chip.RegI != 0 {
		t.Errorf("chip.RegI = 0x%x; want 0x0", chip.RegI)
	}

	if chip.RegSound != 0 {
		t.Errorf("chip.RegSound = 0x%x; want 0x0", chip.RegSound)
	}

	if chip.RegDelay != 0 {
		t.Errorf("chip.RegDelay = 0x%x; want 0x0", chip.RegDelay)
	}
}

////////////////////////////////////////////////////////////////////////////////
// memory read/write functions
////////////////////////////////////////////////////////////////////////////////

func TestReadByte(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Memory[42] = 0xba

	got := chip.ReadByte(42)
	if got != 0xba {
		t.Errorf("chip.ReadByte(42) = 0x%x; want 0xba", got)
	}

	got = chip.ReadByte(65535)
	if got != 0xff {
		t.Errorf("chip.ReadByte(65535) = 0x%x; want 0xff", got)
	}
}

func TestReadShort(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Memory[42] = 0xab
	chip.Memory[43] = 0xcd
	chip.Memory[44] = 0xef

	got := chip.ReadShort(42)
	if got != 0xabcd {
		t.Errorf("chip.ReadShort(42) = 0x%x; want 0xabcd", got)
	}

	got = chip.ReadShort(43)
	if got != 0xcdef {
		t.Errorf("chip.ReadShort(43) = 0x%x; want 0xcdef", got)
	}

	got = chip.ReadShort(65534)
	if got != 0xffff {
		t.Errorf("chip.ReadShort(65534) = 0x%x; want 0xffff", got)
	}
}

func TestWriteByte(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.WriteByte(42, 0xba)

	got := chip.Memory[42]
	if got != 0xba {
		t.Errorf("chip.Memory[42] = 0x%x; want 0xba", got)
	}
}

func TestWriteShort(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.WriteShort(42, 0xabcd)

	got := chip.Memory[42]
	if got != 0xab {
		t.Errorf("chip.Memory[42] = 0x%x; want 0xab", got)
	}

	got = chip.Memory[43]
	if got != 0xcd {
		t.Errorf("chip.Memory[43] = 0x%x; want 0xcd", got)
	}
}

////////////////////////////////////////////////////////////////////////////////
// display read/write functions
////////////////////////////////////////////////////////////////////////////////

func TestDisplayReadByte(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Display[42] = 0xba

	got := chip.ReadDisplayByte(42)
	if got != 0xba {
		t.Errorf("chip.ReadDisplayByte(42) = 0x%x; want 0xba", got)
	}

	got = chip.ReadDisplayByte(65535)
	if got != 0xff {
		t.Errorf("chip.ReadDisplayByte(65535) = 0x%x; want 0xff", got)
	}
}

func TestReadDisplayShort(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.Display[42] = 0xab
	chip.Display[43] = 0xcd
	chip.Display[44] = 0xef

	got := chip.ReadDisplayShort(42)
	if got != 0xabcd {
		t.Errorf("chip.ReadDisplayShort(42) = 0x%x; want 0xabcd", got)
	}

	got = chip.ReadDisplayShort(43)
	if got != 0xcdef {
		t.Errorf("chip.ReadDisplayShort(43) = 0x%x; want 0xcdef", got)
	}

	got = chip.ReadDisplayShort(65534)
	if got != 0xffff {
		t.Errorf("chip.ReadDisplayShort(65534) = 0x%x; want 0xffff", got)
	}
}

func TestDisplayWriteByte(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.WriteDisplayByte(42, 0xba)

	got := chip.Display[42]
	if got != 0xba {
		t.Errorf("chip.Display[42] = 0x%x; want 0xba", got)
	}
}

func TestDisplayWriteShort(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.WriteDisplayShort(42, 0xabcd)

	got := chip.Display[42]
	if got != 0xab {
		t.Errorf("chip.Display[42] = 0x%x; want 0xab", got)
	}

	got = chip.Display[43]
	if got != 0xcd {
		t.Errorf("chip.Display[43] = 0x%x; want 0xcd", got)
	}
}

////////////////////////////////////////////////////////////////////////////////
// stack functions
////////////////////////////////////////////////////////////////////////////////

func TestPushStack(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.StackPtr = chipCfg.SizeStack - 1
	chip.pushStack(0xba)

	if chip.Stack[chipCfg.SizeStack-1] != 0xba {
		t.Errorf("chip.Stack[0x%x] = 0x%x; want 0xba", chipCfg.SizeStack-1, chip.Stack[chipCfg.SizeStack-1])

	}

	if chip.StackPtr != chipCfg.SizeStack {
		t.Errorf("chip.StackPtr = 0x%x; want 0x%x", chip.StackPtr, chipCfg.SizeStack)
	}

	chip.pushStack(0xba)
	if chip.StackPtr != chipCfg.SizeStack {
		t.Errorf("chip.StackPtr = 0x%x; want 0x%x", chip.StackPtr, chipCfg.SizeStack)
	}

}

func TestPopStack(t *testing.T) {
	chipCfg := GetDefaultConfig()
	chip, _ := NewCHIP8(chipCfg)

	chip.StackPtr = 1
	chip.Stack[0] = 0xba

	got := chip.popStack()
	if got != 0xba {
		t.Errorf("chip.popStack() = 0x%x; want 0xba", got)
	}

	if chip.StackPtr != 0x0 {
		t.Errorf("chip.StackPtr = 0x%x; want 0x0", chip.StackPtr)
	}

	got = chip.popStack()
	if got != 0xff {
		t.Errorf("chip.popStack() = 0x%x; want 0xff", got)
	}

	if chip.StackPtr != 0x0 {
		t.Errorf("chip.StackPtr = 0x%x; want 0x0", chip.StackPtr)
	}

}
