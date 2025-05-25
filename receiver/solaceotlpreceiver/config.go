package solaceotlpreceiver

// Config defines configuration for the Solace OTLP receiver
type Config struct {
	Host     string `mapstructure:"host"`     // Solace host/endpoint
	VPN      string `mapstructure:"vpn"`      // Solace VPN name
	Username string `mapstructure:"username"` // Solace username
	Password string `mapstructure:"password"` // Solace password
	Queue    string `mapstructure:"queue"`    // Queue name for receiving messages
}
