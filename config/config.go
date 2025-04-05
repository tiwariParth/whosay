package config

// Config represents application configuration
type Config struct {
	Version string
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		Version: "0.1.0",
	}
}
