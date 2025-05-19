package solaceotlpreceiver

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
)

// Config defines configuration for Solace OTLP receiver.
type Config struct {
	// Endpoint is the Solace broker endpoint
	Endpoint string `mapstructure:"endpoint"`

	// Queue is the name of the queue to consume from
	Queue string `mapstructure:"queue"`

	// Username for Solace authentication
	Username string `mapstructure:"username"`

	// Password for Solace authentication
	Password string `mapstructure:"password"`

	// VPN is the name of the Solace VPN
	VPN string `mapstructure:"vpn"`

	// TLS configuration
	TLS *TLSConfig `mapstructure:"tls"`
}

// TLSConfig defines TLS configuration for Solace connection
type TLSConfig struct {
	// InsecureSkipVerify controls whether to verify the server's certificate
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`

	// CAFile is the path to the CA certificate file
	CAFile string `mapstructure:"ca_file"`

	// CertFile is the path to the client certificate file
	CertFile string `mapstructure:"cert_file"`

	// KeyFile is the path to the client key file
	KeyFile string `mapstructure:"key_file"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("endpoint must be specified")
	}
	if cfg.Queue == "" {
		return fmt.Errorf("queue must be specified")
	}
	return nil
}
