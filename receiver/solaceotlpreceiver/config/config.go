package config

import (
	"fmt"
)

// Config defines configuration for the Solace OTLP receiver
type Config struct {
	// Host is the Solace broker host
	Host string `mapstructure:"host"`
	// Port is the Solace broker port
	Port int `mapstructure:"port"`
	// VPN is the Solace VPN name
	VPN string `mapstructure:"vpn"`
	// Username is the Solace username
	Username string `mapstructure:"username"`
	// Password is the Solace password
	Password string `mapstructure:"password"`
	// Queue is the Solace queue name
	Queue string `mapstructure:"queue"`
	// SSL enables SSL/TLS connection
	SSL bool `mapstructure:"ssl"`
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Host == "" {
		return fmt.Errorf("host must be specified")
	}
	if cfg.Port <= 0 {
		return fmt.Errorf("port must be greater than 0")
	}
	if cfg.VPN == "" {
		return fmt.Errorf("vpn must be specified")
	}
	if cfg.Username == "" {
		return fmt.Errorf("username must be specified")
	}
	if cfg.Password == "" {
		return fmt.Errorf("password must be specified")
	}
	if cfg.Queue == "" {
		return fmt.Errorf("queue must be specified")
	}
	return nil
}
