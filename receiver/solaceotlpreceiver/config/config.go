package config

import (
	"fmt"
)

// Config defines configuration for the Solace OTLP receiver
type Config struct {
	// Queue is the name of the Solace queue to consume from
	Queue string `mapstructure:"queue"`
	// Host is the Solace broker host
	Host string `mapstructure:"host"`
	// VPN is the Solace VPN name
	VPN string `mapstructure:"vpn"`
	// Username is the Solace username
	Username string `mapstructure:"username"`
	// Password is the Solace password
	Password string `mapstructure:"password"`
	// receiver stores the current receiver instance
	receiver interface{}
}

// GetReceiver returns the current receiver instance
func (c *Config) GetReceiver() interface{} {
	return c.receiver
}

// SetReceiver sets the current receiver instance
func (c *Config) SetReceiver(r interface{}) {
	c.receiver = r
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Host == "" {
		return fmt.Errorf("host must be specified")
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
