package solaceotlpreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/config"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/base"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/interfaces"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/logs"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/solace"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/traces"
)

// Receiver is the implementation of the OpenTelemetry receiver for Solace
type Receiver struct {
	baseReceiver   *base.Receiver
	logsReceiver   *logs.Receiver
	tracesReceiver *traces.Receiver
	solaceClient   *solace.Client
}

// NewReceiver creates a new Receiver for Logs and Traces
func NewReceiver(
	settings receiver.Settings,
	config *config.Config,
	logsConsumer consumer.Logs,
	tracesConsumer consumer.Traces,
) *Receiver {
	baseReceiver := base.NewReceiver(settings, config)
	solaceClient := solace.NewClient(settings.TelemetrySettings.Logger, config)

	return &Receiver{
		baseReceiver:   baseReceiver,
		logsReceiver:   logs.NewReceiver(settings, baseReceiver, logsConsumer),
		tracesReceiver: traces.NewReceiver(settings, baseReceiver, tracesConsumer),
		solaceClient:   solaceClient,
	}
}

// Start starts the Receiver
func (r *Receiver) Start(ctx context.Context, host component.Host) error {
	// Connect to Solace
	if err := r.solaceClient.Connect(); err != nil {
		return err
	}

	// Register message handlers
	queueConsumer := r.solaceClient.GetQueueConsumer()
	if queueConsumer != nil {
		queueConsumer.SetMessageListener(r.handleMessage)
	}

	return nil
}

// Shutdown ends the Receiver
func (r *Receiver) Shutdown(ctx context.Context) error {
	return r.solaceClient.Disconnect()
}

// handleMessage processes incoming messages
func (r *Receiver) handleMessage(message interface{}) {
	if msg, ok := message.(interfaces.InboundMessage); ok {
		r.logsReceiver.HandleMessage(msg)
		r.tracesReceiver.HandleMessage(msg)
		return
	}
	r.baseReceiver.GetLogger().Warn("Received message of unknown type")
}
