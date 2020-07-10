package chip8

// Config represents the configuration for the CHIP8 machine
type Config struct {
	ResolutionX, ResolutionY int     // num pixels
	SizeMemory               int     // bytes
	SizeStack                int     // bytes
	SizeDisplay              int     // bytes
	NumRegisters             int     // num 16-bit registers
	ClockFreq                float32 // Hz
	TimerDecrementFreq       int     // Hz
}

// GetDefaultConfig returns the default CHIP8 configuration
func GetDefaultConfig() *Config {
	return &Config{
		ResolutionX:        64,
		ResolutionY:        32,
		SizeMemory:         4096,
		SizeStack:          16,
		SizeDisplay:        256,
		NumRegisters:       16,
		ClockFreq:          500,
		TimerDecrementFreq: 60,
	}
}
