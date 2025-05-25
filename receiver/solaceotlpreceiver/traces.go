package solaceotlpreceiver

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
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
	randNum := rand.Intn(1000000)
	receiver := &TracesReceiver{
		consumer: consumer,
		settings: settings,
		config:   config,
		logger:   settings.TelemetrySettings.Logger,
	}
	receiver.logger.Info("NewTracesReceiver instance created - Build-Check!", zap.Time("created_at", time.Now()), zap.String("queue", config.Queue), zap.Int("rand", randNum))
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
		r.logger.Info("Building QueueConsumer …")
		queueConsumer, err := builder.
			WithMessageListener(r.HandleMessage).
			WithClientName("trace").
			Build(*resource.QueueDurableExclusive(r.config.Queue))
		if err != nil {
			return fmt.Errorf("failed to create queue consumer: %w", err)
		}
		r.logger.Info("MessageListener registered!")
		r.QueueConsumer = queueConsumer
		r.logger.Info("QueueConsumer instance created", zap.Time("created_at", time.Now()), zap.String("queue", r.config.Queue))
		r.logger.Info("Starting QueueConsumer …")
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
		err = receiver.Start()
		if err != nil {
			return fmt.Errorf("failed to start persistent message receiver (SDK): %w", err)
		}
		if regErr := receiver.ReceiveAsync(r.HandleMessage); regErr != nil {
			return fmt.Errorf("failed to register message handler: %w", regErr)
		}
		r.QueueConsumer = receiver
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
	r.logger.Debug("HandleMessage called!")

	r.wg.Add(1)
	defer r.wg.Done()

	payload, ok := msg.GetPayloadAsString()
	r.logger.Debug("Payload received", zap.Bool("ok", ok))
	if !ok {
		r.logger.Error("Failed to get message payload")
		return
	}

	r.logger.Debug("[TEST] Payload content", zap.String("payload", payload))

	// Parse JSON payload
	var traceData struct {
		TraceID      string `json:"trace_id"`
		SpanID       string `json:"span_id"`
		ParentSpanID string `json:"parent_span_id"`
		Name         string `json:"name"`
		Kind         int    `json:"kind"`
		StartTime    int64  `json:"start_time"`
		EndTime      int64  `json:"end_time"`
		Status       struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
	}

	if err := json.Unmarshal([]byte(payload), &traceData); err != nil {
		r.logger.Error("Failed to unmarshal trace data", zap.Error(err))
		return
	}

	// Create OTLP trace
	otlpTraces := ptraceotlp.NewExportRequest()
	traces := otlpTraces.Traces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	span := scopeSpans.Spans().AppendEmpty()

	// Convert IDs
	traceID, err := hexStringToTraceID(traceData.TraceID)
	if err != nil {
		r.logger.Error("Failed to convert trace ID", zap.Error(err))
		return
	}
	spanID, err := hexStringToSpanID(traceData.SpanID)
	if err != nil {
		r.logger.Error("Failed to convert span ID", zap.Error(err))
		return
	}
	parentSpanID, err := hexStringToSpanID(traceData.ParentSpanID)
	if err != nil {
		r.logger.Error("Failed to convert parent span ID", zap.Error(err))
		return
	}

	// Set span data
	span.SetTraceID(traceID)
	span.SetSpanID(spanID)
	span.SetParentSpanID(parentSpanID)
	span.SetName(traceData.Name)
	span.SetKind(ptrace.SpanKind(traceData.Kind))
	span.SetStartTimestamp(pcommon.Timestamp(traceData.StartTime))
	span.SetEndTimestamp(pcommon.Timestamp(traceData.EndTime))
	span.Status().SetCode(ptrace.StatusCode(traceData.Status.Code))
	span.Status().SetMessage(traceData.Status.Message)

	r.logger.Debug("Trace successfully deserialized, forwarding to consumer...")
	if err := r.consumer.ConsumeTraces(context.Background(), traces); err != nil {
		r.logger.Error("Failed to consume traces", zap.Error(err))
		return
	}
	if receiver, ok := r.QueueConsumer.(interface {
		Ack(message.InboundMessage) error
	}); ok {
		if err := receiver.Ack(msg); err != nil {
			r.logger.Error("Failed to acknowledge message", zap.Error(err))
		}
	} else {
		r.logger.Error("QueueConsumer does not support Ack", zap.String("type", fmt.Sprintf("%T", r.QueueConsumer)))
	}
	r.logger.Info("MessageListener registered!")
}

// hexStringToTraceID converts a hex string to a TraceID
func hexStringToTraceID(s string) (pcommon.TraceID, error) {
	var traceID pcommon.TraceID
	_, err := hex.Decode(traceID[:], []byte(s))
	return traceID, err
}

// hexStringToSpanID converts a hex string to a SpanID
func hexStringToSpanID(s string) (pcommon.SpanID, error) {
	var spanID pcommon.SpanID
	_, err := hex.Decode(spanID[:], []byte(s))
	return spanID, err
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
