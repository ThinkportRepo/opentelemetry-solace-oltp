package solace

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"

	solaceotlp "github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/config"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/resource"
)

// Client represents a Solace client
type Client struct {
	logger           *zap.Logger
	config           *solaceotlp.Config
	messagingService interface{}
	queueConsumer    QueueConsumer
	messageListener  MessageListener
	stopChan         chan struct{}
	wg               sync.WaitGroup
}

// NewClient creates a new Solace client
func NewClient(logger *zap.Logger, config *solaceotlp.Config) *Client {
	return &Client{
		logger:   logger,
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// Connect connects the client to Solace
func (c *Client) Connect() error {
	// Build messaging service from config
	props := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                c.config.Host,
		config.ServicePropertyVPNName:                    c.config.VPN,
		config.AuthenticationPropertySchemeBasicUserName: c.config.Username,
		config.AuthenticationPropertySchemeBasicPassword: c.config.Password,
	}

	// Set SSL properties
	props[config.TransportLayerSecurityPropertyTrustStorePath] = c.getTrustStorePath()

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

	// Create queue consumer
	consumerBuilder := service.CreatePersistentMessageReceiverBuilder()
	consumer, err := consumerBuilder.
		WithMessageAutoAcknowledgement().
		Build(resource.QueueDurableExclusive(c.config.Queue))
	if err != nil {
		return fmt.Errorf("failed to build queue consumer: %w", err)
	}

	// Start consumer
	if err := consumer.Start(); err != nil {
		return fmt.Errorf("failed to start queue consumer: %w", err)
	}

	c.queueConsumer = NewPersistentMessageReceiverAdapter(consumer)
	c.StartMessageReceiver()
	return nil
}

// Disconnect disconnects the connection to the Solace server
func (c *Client) Disconnect() error {
	// Signal stop to message receiver
	close(c.stopChan)
	c.wg.Wait()

	if c.queueConsumer != nil {
		if terminator, ok := c.queueConsumer.(interface{ Terminate(time.Duration) error }); ok {
			if err := terminator.Terminate(1 * time.Second); err != nil {
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

// GetQueueConsumer returns the QueueConsumer
func (c *Client) GetQueueConsumer() QueueConsumer {
	return c.queueConsumer
}

// getTrustStorePath returns the trust store path from environment variable or default
func (c *Client) getTrustStorePath() string {
	if path := os.Getenv("SOLACE_TRUST_STORE_PATH"); path != "" {
		return path
	}
	return filepath.Join(os.Getenv("HOME"), ".solace", "DigiCertGlobalRootCA.crt.pem")
}

// SetMessageListener sets the MessageListener for the client
func (c *Client) SetMessageListener(listener MessageListener) {
	c.messageListener = listener
	if c.queueConsumer != nil {
		c.queueConsumer.SetMessageListener(listener)
	}
}

// StartMessageReceiver starts receiving messages
func (c *Client) StartMessageReceiver() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.stopChan:
				return
			default:
				msg, err := c.queueConsumer.ReceiveMessage(1 * time.Second)
				if err != nil {
					if err.Error() != "receiver has been terminated, no messages to receive" &&
						err.Error() != "timed out waiting for message on call to Receive" {
						c.logger.Error("Error receiving message", zap.Error(err))
					}
					continue
				}
				if msg != nil && c.messageListener != nil {
					c.messageListener.OnMessage(NewSolaceInboundMessage(msg))
				}
			}
		}
	}()
}
