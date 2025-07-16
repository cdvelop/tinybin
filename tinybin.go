package tinybin

// Config holds configuration for a TinyBin handler.
type Config struct {
	MaxDepth int // Maximum nesting depth for struct analysis
}

// TinyBin is the main handler for encoding and decoding operations.
type TinyBin struct {
	stObjects []stObject // Internal cache of registered struct schemas
	*Config              // Pointer to configuration
}

// New creates a new TinyBin handler with optional configuration.
// If no config is provided or config is nil, uses default values.
func New(cfg ...*Config) *TinyBin {
	var conf *Config
	if len(cfg) == 0 || cfg[0] == nil {
		conf = &Config{MaxDepth: 8}
	} else {
		conf = cfg[0]
	}
	return &TinyBin{Config: conf}
}
