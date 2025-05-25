package solace

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	myconfig "github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/config"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
)

// Client represents a Solace messaging client
type Client struct {
	logger           *zap.Logger
	config           *myconfig.Config
	messagingService interface{}
	queueConsumer    interface{}
}

// NewClient creates a new Solace client
func NewClient(logger *zap.Logger, config *myconfig.Config) *Client {
	return &Client{
		logger: logger,
		config: config,
	}
}

// Connect connects to the Solace broker
func (c *Client) Connect() error {
	// Build messaging service from config
	props := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                c.config.Host,
		config.ServicePropertyVPNName:                    c.config.VPN,
		config.AuthenticationPropertySchemeBasicUserName: c.config.Username,
		config.AuthenticationPropertySchemeBasicPassword: c.config.Password,
	}

	// Set SSL properties if enabled
	if c.config.SSL {
		props[config.TransportLayerSecurityPropertyTrustStorePath] = c.getTrustStorePath()
	}

	// Create messaging service
	service, err := messaging.NewMessagingServiceBuilder().FromConfigurationProvider(props).Build()
	if err != nil {
		return fmt.Errorf("failed to create messaging service: %w", err)
	}

	// Connect to service
	if err := service.Connect(); err != nil {
		return fmt.Errorf("failed to connect to messaging service: %w", err)
	}

	c.messagingService = service
	return nil
}

// Disconnect disconnects from the Solace broker
func (c *Client) Disconnect() error {
	if c.queueConsumer != nil {
		if terminator, ok := c.queueConsumer.(interface{ Terminate(uint) error }); ok {
			if err := terminator.Terminate(1000); err != nil {
				return fmt.Errorf("failed to terminate queue consumer: %w", err)
			}
		}
	}

	if c.messagingService != nil {
		if disconnector, ok := c.messagingService.(interface{ Disconnect() error }); ok {
			if err := disconnector.Disconnect(); err != nil {
				return fmt.Errorf("failed to disconnect messaging service: %w", err)
			}
		}
	}

	return nil
}

// GetQueueConsumer returns the current queue consumer
func (c *Client) GetQueueConsumer() interface{} {
	return c.queueConsumer
}

// getTrustStorePath returns the trust store path from environment variable or default
func (c *Client) getTrustStorePath() string {
	if path := os.Getenv("SOLACE_TRUST_STORE_PATH"); path != "" {
		return path
	}
	return filepath.Join(os.Getenv("HOME"), ".solace", "truststore.jks")
}
