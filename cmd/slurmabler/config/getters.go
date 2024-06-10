// Package config configures the application on start, exports config, initialization, etc
package config

// GetConfig returns config
func GetConfig() *Config {
	return &cfg
}
