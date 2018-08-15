package p2p

// Config defines configuration the p2p needs.
type Config struct {
	SeedNodes   []string
	ListenAddr  string
	PrivKeyPath string
	ChainID     uint32
	Version     uint16
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		ListenAddr:  "0.0.0.0:6666",
		PrivKeyPath: "priv.key",
		ChainID:     404,
		Version:     1,
	}
}
