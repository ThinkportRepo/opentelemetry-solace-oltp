package solaceotlpreceiver

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
)

type tracesReceiver struct {
	config           *Config
	consumer         consumer.Traces
	logger           *zap.Logger
	settings         receiver.CreateSettings
	messagingService solace.MessagingService
	queueConsumer    solace.QueueConsumer
	wg               sync.WaitGroup
}

func newTracesReceiver(
	settings receiver.CreateSettings,
	config *Config,
	consumer consumer.Traces,
) (receiver.Traces, error) {
	return &tracesReceiver{
		config:   config,
		consumer: consumer,
		logger:   settings.Logger,
		settings: settings,
	}, nil
}

func (r *tracesReceiver) Start(ctx context.Context, host component.Host) error {
	r.logger.Info("Starting Solace OTLP traces receiver",
		zap.String("endpoint", r.config.Endpoint),
		zap.String("queue", r.config.Queue))

	// Create messaging service
	messagingService, err := messaging.NewMessagingServiceBuilder().
		FromConfigurationProvider(config.ServicePropertyMap{
			config.TransportLayerPropertyHost:                r.config.Endpoint,
			config.ServicePropertyVPNName:                    "default",
			config.AuthenticationPropertySchemeBasicUserName: r.config.Username,
			config.AuthenticationPropertySchemeBasicPassword: r.config.Password,
		}).
		Build()
	if err != nil {
		return fmt.Errorf("failed to create messaging service: %w", err)
	}
	r.messagingService = messagingService

	// Connect to Solace
	if err := r.messagingService.Connect(); err != nil {
		return fmt.Errorf("failed to connect to Solace: %w", err)
	}

	// Create queue consumer
	queueConsumer, err := r.messagingService.CreateQueueConsumerBuilder().
		WithMessageAutoAcknowledgement().
		WithMessageListener(r.handleMessage).
		Build(resource.Queue(r.config.Queue))
	if err != nil {
		return fmt.Errorf("failed to create queue consumer: %w", err)
	}
	r.queueConsumer = queueConsumer

	// Start consuming
	if err := r.queueConsumer.Start(); err != nil {
		return fmt.Errorf("failed to start queue consumer: %w", err)
	}

	return nil
}

func (r *tracesReceiver) handleMessage(msg message.InboundMessage) {
	r.wg.Add(1)
	defer r.wg.Done()

	traces, err := parseOTLPTraceMessage(msg)
	if err != nil {
		r.logger.Error("Failed to parse OTLP trace message", zap.Error(err))
		return
	}

	if err := r.consumer.ConsumeTraces(context.Background(), traces); err != nil {
		r.logger.Error("Failed to consume traces", zap.Error(err))
	}
}

func (r *tracesReceiver) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down Solace OTLP traces receiver")

	if r.queueConsumer != nil {
		if err := r.queueConsumer.Terminate(10); err != nil {
			r.logger.Error("Failed to terminate queue consumer", zap.Error(err))
		}
	}

	if r.messagingService != nil {
		if err := r.messagingService.Disconnect(); err != nil {
			r.logger.Error("Failed to disconnect messaging service", zap.Error(err))
		}
	}

	r.wg.Wait()
	return nil
}
