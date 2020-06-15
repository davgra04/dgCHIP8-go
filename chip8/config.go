package chip8

// Config represents the configuration for the CHIP8 machine
type Config struct {
	ResolutionX, ResolutionY int // num pixels
	SizeProgram              int // bytes
	SizeStack                int // bytes
	SizeDisplay              int // bytes
	NumRegisters             int // num 16-bit registers
	ClockFreq                int // Hz
	ScreenRefreshFreq        int // Hz
	TimerDecrementFreq       int // Hz
}

// GetDefaultConfig returns the default CHIP8 configuration
func GetDefaultConfig() *Config {
	return &Config{
		ResolutionX:        64,
		ResolutionY:        32,
		SizeProgram:        4000,
		SizeStack:          16,
		SizeDisplay:        256,
		NumRegisters:       16,
		ClockFreq:          500,
		ScreenRefreshFreq:  60,
		TimerDecrementFreq: 60,
	}
}
