package solaceotlpreceiver

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/mocks"
)

// TracesReceiver implementiert den Receiver für Traces
type TracesReceiver struct {
	consumer         consumer.Traces
	settings         receiver.CreateSettings
	config           *Config
	logger           *zap.Logger
	wg               sync.WaitGroup
	messagingService interface{} // echtes SDK oder Mock
	QueueConsumer    interface{} // speichert den verwendeten QueueConsumer
}

// NewTracesReceiver erstellt einen neuen TracesReceiver
func NewTracesReceiver(settings receiver.CreateSettings, config *Config, consumer consumer.Traces, opts ...interface{}) (*TracesReceiver, error) {
	receiver := &TracesReceiver{
		consumer: consumer,
		settings: settings,
		config:   config,
		logger:   settings.TelemetrySettings.Logger,
	}
	if len(opts) > 0 {
		receiver.messagingService = opts[0]
	}
	return receiver, nil
}

// Start startet den Receiver
func (r *TracesReceiver) Start(ctx context.Context, host component.Host) error {
	r.logger.Info("Starting Solace OTLP traces receiver",
		zap.String("endpoint", r.config.Endpoint),
		zap.String("queue", r.config.Queue))

	// MessagingService initialisieren (SDK oder Mock)
	if r.messagingService == nil {
		ms, err := messaging.NewMessagingServiceBuilder().
			FromConfigurationProvider(config.ServicePropertyMap{
				config.TransportLayerPropertyHost:                r.config.Endpoint,
				config.ServicePropertyVPNName:                    r.config.VPN,
				config.AuthenticationPropertySchemeBasicUserName: r.config.Username,
				config.AuthenticationPropertySchemeBasicPassword: r.config.Password,
			}).
			Build()
		if err != nil {
			return fmt.Errorf("failed to create messaging service: %w", err)
		}
		r.messagingService = ms
	}

	var err error
	type queueConsumerBuilderIface interface {
		WithMessageAutoAcknowledgement() queueConsumerBuilderIface
		WithMessageListener(func(message.InboundMessage)) queueConsumerBuilderIface
		Build(resource.Queue) (interface{ Start() error }, error)
	}
	switch ms := r.messagingService.(type) {
	case interface {
		Connect() error
		CreateQueueConsumerBuilder() interface{}
	}:
		err = ms.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to Solace: %w", err)
		}
		builderIface := ms.CreateQueueConsumerBuilder()
		builder, ok := builderIface.(queueConsumerBuilderIface)
		if !ok {
			return fmt.Errorf("queue consumer builder does not implement required interface")
		}
		queueConsumer, err := builder.
			WithMessageAutoAcknowledgement().
			WithMessageListener(r.HandleMessage).
			Build(*resource.QueueDurableExclusive(r.config.Queue))
		if err != nil {
			return fmt.Errorf("failed to create queue consumer: %w", err)
		}
		r.QueueConsumer = queueConsumer
		err = queueConsumer.Start()
		if err != nil {
			return fmt.Errorf("failed to start queue consumer: %w", err)
		}
	case mocks.MessagingService:
		err = ms.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to Solace (mock): %w", err)
		}
		queueConsumerBuilder := ms.CreateQueueConsumerBuilder()
		queueConsumer, err := queueConsumerBuilder.
			WithMessageAutoAcknowledgement().
			WithMessageListener(r.HandleMessage).
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
	default:
		return fmt.Errorf("unsupported messagingService type")
	}

	return nil
}

// Shutdown beendet den Receiver
func (r *TracesReceiver) Shutdown(ctx context.Context) error {
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

// HandleMessage verarbeitet eine eingehende Nachricht
func (r *TracesReceiver) HandleMessage(msg message.InboundMessage) {
	r.wg.Add(1)
	defer r.wg.Done()

	payload, ok := msg.GetPayloadAsBytes()
	if !ok {
		r.logger.Error("Failed to get message payload")
		return
	}

	otlpTraces := ptraceotlp.NewExportRequest()
	if err := otlpTraces.UnmarshalProto(payload); err != nil {
		r.logger.Error("Failed to unmarshal traces", zap.Error(err))
		return
	}

	if err := r.consumer.ConsumeTraces(context.Background(), otlpTraces.Traces()); err != nil {
		r.logger.Error("Failed to consume traces", zap.Error(err))
	}
}

// GetVPN returns the VPN configuration
func (r *TracesReceiver) GetVPN() string {
	return r.config.VPN
}
