package solaceotlpreceiver

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/mocks"
	"go.opentelemetry.io/collector/receiver"
)

// TracesReceiver implements the receiver for traces
type TracesReceiver struct {
	consumer         consumer.Traces
	settings         receiver.Settings
	config           *Config
	logger           *zap.Logger
	wg               sync.WaitGroup
	messagingService interface{} // real SDK or mock
	QueueConsumer    interface{} // stores the used QueueConsumer
}

// NewTracesReceiver creates a new TracesReceiver
func NewTracesReceiver(
	settings receiver.Settings,
	config *Config,
	consumer consumer.Traces,
	opts ...interface{},
) (*TracesReceiver, error) {
	receiver := &TracesReceiver{
		consumer: consumer,
		settings: settings,
		config:   config,
		logger:   settings.TelemetrySettings.Logger,
	}
	receiver.logger.Info("NewTracesReceiver instance created", zap.Time("created_at", time.Now()), zap.String("queue", config.Queue))
	if len(opts) > 0 {
		receiver.messagingService = opts[0]
	}
	return receiver, nil
}

// Start starts the receiver
func (r *TracesReceiver) Start(ctx context.Context, host component.Host) error {
	r.logger.Info("Start() called for TracesReceiver", zap.Time("start_time", time.Now()), zap.String("queue", r.config.Queue))

	r.logger.Info("Starting Solace OTLP traces receiver",
		zap.String("endpoint", r.config.Endpoint),
		zap.String("queue", r.config.Queue))

	// MessagingService initialisieren (SDK oder Mock)
	if r.messagingService == nil {
		r.logger.Info("Creating new MessagingService (SDK or Mock)")
		ms, err := messaging.NewMessagingServiceBuilder().
			FromConfigurationProvider(config.ServicePropertyMap{
				config.TransportLayerPropertyHost:                   r.config.Endpoint,
				config.ServicePropertyVPNName:                       r.config.VPN,
				config.AuthenticationPropertySchemeBasicUserName:    r.config.Username,
				config.AuthenticationPropertySchemeBasicPassword:    r.config.Password,
				config.TransportLayerSecurityPropertyTrustStorePath: getTrustStorePath(),
			}).
			WithTransportSecurityStrategy(
				config.NewTransportSecurityStrategy().WithCertificateValidation(true, false, "", ""),
			).
			Build()
		if err != nil {
			return fmt.Errorf("failed to create messaging service: %w", err)
		}
		r.messagingService = ms
	}

	r.logger.Info("MessagingService instance type", zap.String("type", fmt.Sprintf("%T", r.messagingService)))

	r.logger.Info("Initializing QueueConsumer...")

	var err error
	type queueConsumerBuilderIface interface {
		WithMessageAutoAcknowledgement() queueConsumerBuilderIface
		WithMessageListener(func(message.InboundMessage)) queueConsumerBuilderIface
		WithClientName(string) queueConsumerBuilderIface
		Build(resource.Queue) (interface{ Start() error }, error)
	}
	switch ms := r.messagingService.(type) {
	case interface {
		Connect() error
		CreateQueueConsumerBuilder() interface{}
	}:
		r.logger.Info("Using generic MessagingService interface")
		err = ms.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to Solace: %w", err)
		}
		builderIface := ms.CreateQueueConsumerBuilder()
		builder, ok := builderIface.(queueConsumerBuilderIface)
		if !ok {
			return fmt.Errorf("queue consumer builder does not implement required interface")
		}
		r.logger.Info("Building QueueConsumer...")
		queueConsumer, err := builder.
			WithMessageAutoAcknowledgement().
			WithMessageListener(r.HandleMessage).
			WithClientName("trace").
			Build(*resource.QueueDurableExclusive(r.config.Queue))
		if err != nil {
			return fmt.Errorf("failed to create queue consumer: %w", err)
		}
		r.QueueConsumer = queueConsumer
		r.logger.Info("QueueConsumer instance created", zap.Time("created_at", time.Now()), zap.String("queue", r.config.Queue))
		r.logger.Info("Starting QueueConsumer...")
		err = queueConsumer.Start()
		if err != nil {
			return fmt.Errorf("failed to start queue consumer: %w", err)
		}
		r.logger.Info("QueueConsumer started", zap.Time("started_at", time.Now()), zap.String("queue", r.config.Queue))
	case mocks.MessagingService:
		r.logger.Info("Using Mock MessagingService")
		err = ms.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to Solace (mock): %w", err)
		}
		queueConsumerBuilder := ms.CreateQueueConsumerBuilder()
		queueConsumer, err := queueConsumerBuilder.
			WithMessageAutoAcknowledgement().
			WithMessageListener(r.HandleMessage).
			WithClientName("trace-consumer").
			Build(*resource.QueueDurableExclusive(r.config.Queue))
		if err != nil {
			return fmt.Errorf("failed to create queue consumer (mock): %w", err)
		}
		if starter, ok := queueConsumer.(interface{ Start() error }); ok {
			err = starter.Start()
			if err != nil {
				return fmt.Errorf("failed to start queue consumer (mock): %w", err)
			}
		}
		r.QueueConsumer = queueConsumer
	case solace.MessagingService:
		r.logger.Info("Using real Solace SDK MessagingService")
		err = ms.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to Solace (SDK): %w", err)
		}
		builder := ms.CreatePersistentMessageReceiverBuilder()
		receiver, err := builder.Build(resource.QueueDurableExclusive(r.config.Queue))
		if err != nil {
			return fmt.Errorf("failed to build persistent message receiver (SDK): %w", err)
		}
		r.QueueConsumer = receiver
		r.logger.Info("Starting persistent message receiver (SDK)...")
		err = receiver.Start()
		if err != nil {
			return fmt.Errorf("failed to start persistent message receiver (SDK): %w", err)
		}
	default:
		r.logger.Error("Unknown MessagingService type!", zap.Any("type", ms))
		return fmt.Errorf("unsupported messagingService type")
	}

	r.logger.Info("Solace OTLP traces receiver started successfully!")

	return nil
}

// Shutdown stops the receiver
func (r *TracesReceiver) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutdown() called for TracesReceiver", zap.Time("shutdown_time", time.Now()), zap.String("queue", r.config.Queue))
	r.logger.Info("Shutting down Solace OTLP traces receiver")
	// Hier ggf. weitere Aufräumarbeiten, z.B. Disconnect
	if r.QueueConsumer != nil {
		if terminator, ok := r.QueueConsumer.(interface{ Terminate(uint) error }); ok {
			_ = terminator.Terminate(10)
		}
	}
	if r.messagingService != nil {
		if disconnector, ok := r.messagingService.(interface{ Disconnect() error }); ok {
			_ = disconnector.Disconnect()
		}
	}
	return nil
}

// HandleMessage processes an incoming message
func (r *TracesReceiver) HandleMessage(msg message.InboundMessage) {
	r.logger.Info("HandleMessage called - new message received!")
	r.wg.Add(1)
	defer r.wg.Done()

	payload, ok := msg.GetPayloadAsBytes()
	if !ok {
		r.logger.Error("Failed to get message payload")
		msg.Dispose()
		return
	}

	r.logger.Info("Message received, attempting to deserialize...", zap.Int("payload_len", len(payload)))

	otlpTraces := ptraceotlp.NewExportRequest()
	if err := otlpTraces.UnmarshalProto(payload); err != nil {
		r.logger.Error("Failed to unmarshal traces", zap.Error(err))
		msg.Dispose()
		return
	}

	r.logger.Info("Trace successfully deserialized, forwarding to consumer...")

	if err := r.consumer.ConsumeTraces(context.Background(), otlpTraces.Traces()); err != nil {
		r.logger.Error("Failed to consume traces", zap.Error(err))
		msg.Dispose()
		return
	}

	// Nach erfolgreicher Verarbeitung die Nachricht bestätigen
	msg.Dispose()
	r.logger.Info("Message successfully acknowledged!")
}

// GetVPN returns the VPN configuration
func (r *TracesReceiver) GetVPN() string {
	return r.config.VPN
}

// getTrustStorePath returns the truststore path from environment variable or default
func getTrustStorePath() string {
	if path := os.Getenv("SESSION_SSL_TRUST_STORE_DIR"); path != "" {
		return path
	}
	return "truststore"
}
