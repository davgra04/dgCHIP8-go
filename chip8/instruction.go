package chip8

import (
	"fmt"
	"math/rand"
)

////////////////////////////////////////////////////////////////////////////////
// decode and execute
////////////////////////////////////////////////////////////////////////////////

func (chip *CHIP8) decodeAndExecuteInstruction(instruction uint16) {

	InstNibble := []uint8{
		uint8((instruction >> 12) & 0xf),
		uint8((instruction >> 8) & 0xf),
		uint8((instruction >> 4) & 0xf),
		uint8(instruction & 0xf),
	}

	InstByte := []uint16{
		uint16((instruction >> 8) & 0xff),
		uint16(instruction & 0xff),
	}

	switch InstNibble[0] {
	case 0x0:
		switch instruction {
		case 0x00e0:
			chip.instructionClearScreen()
			break
		case 0x00ee:
			chip.instructionReturnSubroutine()
		}
		break
	case 0x1:
		chip.instructionJump(instruction)
		break
	case 0x2:
		chip.instructionCallSubroutine(instruction)
		break
	case 0x3:
		chip.instructionSkipEqualByte(instruction)
		break
	case 0x4:
		chip.instructionSkipNotEqualByte(instruction)
		break
	case 0x5:
		chip.instructionSkipEqualReg(instruction)
		break
	case 0x6:
		chip.instructionLoadByte(instruction)
		break
	case 0x7:
		chip.instructionAddByte(instruction)
		break
	case 0x8:
		switch InstNibble[3] {
		case 0x0:
			chip.instructionLoadReg(instruction)
			break
		case 0x1:
			chip.instructionOr(instruction)
			break
		case 0x2:
			chip.instructionAnd(instruction)
			break
		case 0x3:
			chip.instructionXor(instruction)
			break
		case 0x4:
			chip.instructionAddReg(instruction)
			break
		case 0x5:
			chip.instructionSubReg(instruction)
			break
		case 0x6:
			chip.instructionShiftRight(instruction)
			break
		case 0x7:
			chip.instructionSubNReg(instruction)
			break
		case 0xe:
			chip.instructionShiftLeft(instruction)
			break
		}
		break
	case 0x9:
		chip.instructionSkipNotEqualReg(instruction)
		break
	case 0xa:
		chip.instructionLoadRegI(instruction)
		break
	case 0xb:
		chip.instructionJumpReg(instruction)
		break
	case 0xc:
		chip.instructionRand(instruction)
		break
	case 0xd:
		chip.instructionDrawSprite(instruction)
		break
	case 0xe:
		switch InstByte[1] {
		case 0x9e:
			chip.instructionSkipKey(instruction)
			break
		case 0xa1:
			chip.instructionSkipNotKey(instruction)
			break
		default:
			break
		}
		break
	case 0xf:
		switch InstByte[1] {
		case 0x07:
			chip.instructionReadDelayTimer(instruction)
			break
		case 0x0A:
			chip.instructionWaitForKey(instruction)
			break
		case 0x15:
			chip.instructionSetDelayTimer(instruction)
			break
		case 0x18:
			chip.instructionSetSoundTimer(instruction)
			break
		case 0x1e:
			chip.instructionAddRegI(instruction)
			break
		case 0x29:
			chip.instructionLoadSprite(instruction)
			break
		case 0x33:
			chip.instructionLoadBCD(instruction)
			break
		case 0x55:
			chip.instructionLoadMulti(instruction)
			break
		case 0x65:
			chip.instructionReadMulti(instruction)
			break
		default:
			fmt.Printf("Invalid instruction 0x%x\n", instruction)
			break
		}
		break
	default:
		fmt.Printf("Invalid instruction 0x%x\n", instruction)
		break
	}

}

////////////////////////////////////////////////////////////////////////////////
// instructions
////////////////////////////////////////////////////////////////////////////////

// 00E0 - CLS
// Clear the display.
func (chip *CHIP8) instructionClearScreen() {
	chip.clearDisplay()
}

// 00EE - RET
// Return from a subroutine.
func (chip *CHIP8) instructionReturnSubroutine() {
	chip.PC = chip.popStack()
}

// 1nnn - JP addr
// Jump to location nnn.
func (chip *CHIP8) instructionJump(instruction uint16) {
	addr := instruction & 0xfff
	chip.PC = addr
}

// 2nnn - CALL addr
// Call subroutine at nnn.
func (chip *CHIP8) instructionCallSubroutine(instruction uint16) {
	addr := instruction & 0xfff

	if addr < 0x200 {
		return
	}

	chip.pushStack(chip.MAR)
	chip.PC = addr
}

// 3xkk - SE Vx, byte
// Skip next instruction if Vx = kk.
func (chip *CHIP8) instructionSkipEqualByte(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	value := uint8(instruction & 0xff)
	if chip.Reg[regIdx] == value {
		chip.PC += 2
	}
}

// 4xkk - SNE Vx, byte
// Skip next instruction if Vx != kk.
func (chip *CHIP8) instructionSkipNotEqualByte(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	value := uint8(instruction & 0xff)
	if chip.Reg[regIdx] != value {
		chip.PC += 2
	}
}

// 5xy0 - SE Vx, Vy
// Skip next instruction if Vx = Vy.
func (chip *CHIP8) instructionSkipEqualReg(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf
	if chip.Reg[regXIdx] == chip.Reg[regYIdx] {
		chip.PC += 2
	}
}

// 6xkk - LD Vx, byte
// Set Vx = kk.
func (chip *CHIP8) instructionLoadByte(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	value := uint8(instruction & 0xff)
	chip.Reg[regIdx] = value
}

// 7xkk - ADD Vx, byte
// Set Vx = Vx + kk.
func (chip *CHIP8) instructionAddByte(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	value := uint16(instruction & 0xff)
	sum := uint16(chip.Reg[regIdx]) + value

	if sum > 0xff {
		chip.Reg[0xf] = 0x1
	} else {
		chip.Reg[0xf] = 0x0
	}

	chip.Reg[regIdx] = uint8(sum & 0xff)
}

// 8xy0 - LD Vx, Vy
// Set Vx = Vy.
func (chip *CHIP8) instructionLoadReg(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf
	chip.Reg[regXIdx] = chip.Reg[regYIdx]
}

// 8xy1 - OR Vx, Vy
// Set Vx = Vx OR Vy.
func (chip *CHIP8) instructionOr(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf
	chip.Reg[regXIdx] = chip.Reg[regXIdx] | chip.Reg[regYIdx]
}

// 8xy2 - AND Vx, Vy
// Set Vx = Vx AND Vy.
func (chip *CHIP8) instructionAnd(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf
	chip.Reg[regXIdx] = chip.Reg[regXIdx] & chip.Reg[regYIdx]
}

// 8xy3 - XOR Vx, Vy
// Set Vx = Vx XOR Vy.
func (chip *CHIP8) instructionXor(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf
	chip.Reg[regXIdx] = chip.Reg[regXIdx] ^ chip.Reg[regYIdx]
}

// 8xy4 - ADD Vx, Vy
// Set Vx = Vx + Vy, set VF = carry.
func (chip *CHIP8) instructionAddReg(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf
	sum := uint16(chip.Reg[regXIdx]) + uint16(chip.Reg[regYIdx])

	if sum > 0xff {
		chip.Reg[0xf] = 0x1
	} else {
		chip.Reg[0xf] = 0x0
	}

	chip.Reg[regXIdx] = uint8(sum & 0xff)
}

// 8xy5 - SUB Vx, Vy
// Set Vx = Vx - Vy, set VF = NOT borrow.
func (chip *CHIP8) instructionSubReg(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf

	if chip.Reg[regXIdx] > chip.Reg[regYIdx] {
		chip.Reg[regXIdx] = chip.Reg[regXIdx] - chip.Reg[regYIdx]
		chip.Reg[0xf] = 0x1
	} else {
		sub := 0x100 + uint16(chip.Reg[regXIdx]) - uint16(chip.Reg[regYIdx])
		chip.Reg[regXIdx] = uint8(sub & 0xff)
		chip.Reg[0xf] = 0x0

	}
}

// 8xy6 - SHR Vx
// Set Vx = Vx SHR 1.
func (chip *CHIP8) instructionShiftRight(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	if chip.Reg[regIdx]&0x1 == 0x1 {
		chip.Reg[0xf] = 0x1
	} else {
		chip.Reg[0xf] = 0x0
	}

	chip.Reg[regIdx] = chip.Reg[regIdx] >> 1
}

// 8xy7 - SUBN Vx, Vy
// Set Vx = Vy - Vx, set VF = NOT borrow.
func (chip *CHIP8) instructionSubNReg(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf

	if chip.Reg[regXIdx] < chip.Reg[regYIdx] {
		chip.Reg[regXIdx] = chip.Reg[regYIdx] - chip.Reg[regXIdx]
		chip.Reg[0xf] = 0x1
	} else {
		sub := 0x100 + uint16(chip.Reg[regYIdx]) - uint16(chip.Reg[regXIdx])
		chip.Reg[regXIdx] = uint8(sub & 0xff)
		chip.Reg[0xf] = 0x0

	}
}

// 8xyE - SHL Vx {, Vy}
// Set Vx = Vx SHL 1.
func (chip *CHIP8) instructionShiftLeft(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	if chip.Reg[regIdx]&0x80 == 0x80 {
		chip.Reg[0xf] = 0x1
	} else {
		chip.Reg[0xf] = 0x0
	}

	chip.Reg[regIdx] = chip.Reg[regIdx] << 1
}

// 9xy0 - SNE Vx, Vy
// Skip next instruction if Vx != Vy.
func (chip *CHIP8) instructionSkipNotEqualReg(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf
	if chip.Reg[regXIdx] != chip.Reg[regYIdx] {
		chip.PC += 2
	}
}

// Annn - LD I, addr
// Set I = nnn.
func (chip *CHIP8) instructionLoadRegI(instruction uint16) {
	value := uint16(instruction & 0xfff)
	chip.RegI = value
}

// Bnnn - JP V0, addr
// Jump to location nnn + V0.
func (chip *CHIP8) instructionJumpReg(instruction uint16) {
	value := uint16(instruction & 0xfff)
	chip.PC = value + uint16(chip.Reg[0x0])
}

// Cxkk - RND Vx, byte
// Set Vx = random byte AND kk.
func (chip *CHIP8) instructionRand(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	value := uint8(instruction & 0xff)
	r := uint8(rand.Int())
	chip.Reg[regIdx] = r & value
}

// Dxyn - DRW Vx, Vy, nibble
// Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
func (chip *CHIP8) instructionDrawSprite(instruction uint16) {
	regXIdx := instruction >> 8 & 0xf
	regYIdx := instruction >> 4 & 0xf
	x := uint16(chip.Reg[regXIdx])
	y := uint16(chip.Reg[regYIdx])
	bytes := uint8(instruction & 0xf)

	if chip.Cfg.DrawWrap {
		chip.drawSpriteWrap(x, y, bytes)
	} else {
		chip.drawSpriteNoWrap(x, y, bytes)
	}

}

func (chip *CHIP8) drawSpriteNoWrap(x, y uint16, bytes uint8) {
	collision := false

	for i := uint16(0); i < uint16(bytes); i++ {

		// skip drawing bytes past the bottom of the screen
		if y+i >= uint16(chip.Cfg.ResolutionY) {
			continue
		}

		// skip drawing bytes past the right of the screen
		if x >= uint16(chip.Cfg.ResolutionX) {
			continue
		}

		// prepare shortToDraw, get current display address and contents
		spriteByte := chip.ReadByte(chip.RegI + i)
		shortToDraw := uint16(spriteByte) << 8
		shortToDraw = shortToDraw >> (x % 8)

		curDisplayAddr := ((y+i)*uint16(chip.Cfg.ResolutionX) + x) / 8
		curDisplayShort := chip.ReadDisplayShort(curDisplayAddr)

		if x >= uint16(chip.Cfg.ResolutionX-8) {
			// only draw 1st byte if 2nd byte is off screen
			byteToDraw := uint8((shortToDraw >> 8) & 0xff)
			curDisplayByte := chip.ReadDisplayByte(curDisplayAddr)

			if byteToDraw&curDisplayByte != 0 {
				collision = true
			}
			chip.WriteDisplayByte(curDisplayAddr, byteToDraw^curDisplayByte)

		} else {
			// draw full short
			if shortToDraw&curDisplayShort != 0 {
				collision = true
			}
			chip.WriteDisplayShort(curDisplayAddr, shortToDraw^curDisplayShort)
		}

	}

	if collision {
		chip.Reg[0xf] = 0x1
	} else {
		chip.Reg[0xf] = 0x0
	}
}

func (chip *CHIP8) drawSpriteWrap(x, y uint16, bytes uint8) {
	collision := false

	var xAdjusted, yAdjusted uint16

	for i := uint16(0); i < uint16(bytes); i++ {

		yAdjusted = y
		xAdjusted = x

		// adjust y value for bytes past the bottom of the screen
		for yAdjusted+i >= uint16(chip.Cfg.ResolutionY) {
			yAdjusted -= uint16(chip.Cfg.ResolutionY)
		}

		// adjust x value for bytes past the right of the screen
		for xAdjusted >= uint16(chip.Cfg.ResolutionX) {
			xAdjusted -= uint16(chip.Cfg.ResolutionX)
		}

		// prepare shortToDraw, get current display address and contents
		spriteByte := chip.ReadByte(chip.RegI + i)
		shortToDraw := uint16(spriteByte) << 8
		shortToDraw = shortToDraw >> (xAdjusted % 8)

		curDisplayAddr := ((yAdjusted+i)*uint16(chip.Cfg.ResolutionX) + xAdjusted) / 8
		curDisplayShort := chip.ReadDisplayShort(curDisplayAddr)

		if xAdjusted >= uint16(chip.Cfg.ResolutionX-8) {
			// need to wrap 2nd byte

			byteToDraw1 := uint8((shortToDraw >> 8) & 0xff)
			byteToDraw2 := uint8(shortToDraw & 0xff)

			curDisplayAddr1 := curDisplayAddr
			curDisplayAddr2 := curDisplayAddr + 1 - uint16(chip.Cfg.ResolutionX/8)

			curDisplayByte1 := chip.ReadDisplayByte(curDisplayAddr1)
			curDisplayByte2 := chip.ReadDisplayByte(curDisplayAddr2)

			if (byteToDraw1&curDisplayByte1)|(byteToDraw2&curDisplayByte2) != 0 {
				collision = true
			}

			chip.WriteDisplayByte(curDisplayAddr1, byteToDraw1^curDisplayByte1)
			chip.WriteDisplayByte(curDisplayAddr2, byteToDraw2^curDisplayByte2)

		} else {
			// no need to wrap horizontally
			if shortToDraw&curDisplayShort != 0 {
				collision = true
			}

			chip.WriteDisplayShort(curDisplayAddr, shortToDraw^curDisplayShort)
		}

	}

	if collision {
		chip.Reg[0xf] = 0x1
	} else {
		chip.Reg[0xf] = 0x0
	}
}

// Ex9E - SKP Vx
// Skip next instruction if key with the value of Vx is pressed.
func (chip *CHIP8) instructionSkipKey(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	if chip.Keys[chip.Reg[regIdx]] {
		chip.PC += 2
	}
}

// ExA1 - SKNP Vx
// Skip next instruction if key with the value of Vx is not pressed.
func (chip *CHIP8) instructionSkipNotKey(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	if !chip.Keys[chip.Reg[regIdx]] {
		chip.PC += 2
	}
}

// Fx07 - LD Vx, DT
// Set Vx = delay timer value.
func (chip *CHIP8) instructionReadDelayTimer(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	chip.Reg[regIdx] = chip.RegDelay
}

// Fx0A - LD Vx, K
// Wait for a key press, store the value of the key in Vx.
func (chip *CHIP8) instructionWaitForKey(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	if !chip.watchingKeys {
		// start watching for keys, initialize KeysPrev
		chip.watchingKeys = true
		for i := range chip.Keys {
			chip.KeysPrev[i] = chip.Keys[i]
		}
	}

	for i := 0; i < len(chip.Keys); i++ {
		if chip.Keys[i] != chip.KeysPrev[i] {
			if chip.Keys[i] {
				// we got our key, it is a keydown event
				chip.Reg[regIdx] = uint8(i)
				chip.watchingKeys = false
				return
			}
		}
	}

	// key not pressed, reset PC and update keyPrev
	chip.PC = chip.MAR
	for i := range chip.Keys {
		chip.KeysPrev[i] = chip.Keys[i]
	}
}

// Fx15 - LD DT, Vx
// Set delay timer = Vx.
func (chip *CHIP8) instructionSetDelayTimer(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	chip.RegDelay = chip.Reg[regIdx]
}

// Fx18 - LD ST, Vx
// Set sound timer = Vx.
func (chip *CHIP8) instructionSetSoundTimer(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	if chip.Reg[regIdx] > 0 {
		chip.RegSound = chip.Reg[regIdx]
		// TODO: START BEEP
	}
}

// Fx1E - ADD I, Vx
// Set I = I + Vx.
func (chip *CHIP8) instructionAddRegI(instruction uint16) {
	regIdx := instruction >> 8 & 0xf
	chip.RegI += uint16(chip.Reg[regIdx])
}

// Fx29 - LD F, Vx
// Set I = location of sprite for digit Vx.
func (chip *CHIP8) instructionLoadSprite(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	if chip.Reg[regIdx] < 16 {
		chip.RegI = uint16(chip.Reg[regIdx]) * 5
	}
}

// Fx33 - LD B, Vx
// Store BCD representation of Vx in memory locations I, I+1, and I+2.
func (chip *CHIP8) instructionLoadBCD(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	// fmt.Printf("Writing Reg%x to 0x%x, 0x%x, and 0x%x\n", regIdx, chip.RegI, chip.RegI+1, chip.RegI+2)

	chip.Memory[chip.RegI] = (chip.Reg[regIdx] / 100) % 10
	chip.Memory[chip.RegI+1] = (chip.Reg[regIdx] / 10) % 10
	chip.Memory[chip.RegI+2] = chip.Reg[regIdx] % 10
}

// Fx55 - LD [I], Vx
// Store registers V0 through Vx in memory starting at location I.
func (chip *CHIP8) instructionLoadMulti(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	for i := uint16(0); i <= regIdx; i++ {
		chip.Memory[chip.RegI+i] = chip.Reg[i]
	}
}

// Fx65 - LD Vx, [I]
// Read registers V0 through Vx from memory starting at location I.
func (chip *CHIP8) instructionReadMulti(instruction uint16) {
	regIdx := instruction >> 8 & 0xf

	for i := uint16(0); i <= regIdx; i++ {
		chip.Reg[i] = chip.Memory[chip.RegI+i]
	}
}
