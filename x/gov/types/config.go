package types

// Config is a config struct used for intialising the gov module to avoid using globals.
type Config struct {
	// MaxMetadataLen defines the maximum proposal metadata length.
	MaxMetadataLen uint64
	// MaxOptionsLen defines the maximum number of options a proposal can have.
	MaxOptionsLen uint64
}

// DefaultConfig returns the default config for gov.
func DefaultConfig() Config {
	return Config{
		MaxMetadataLen: 255,
		MaxOptionsLen:  1,
	}
}
