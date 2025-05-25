package solaceotlpreceiver

import (
	"context"
	"encoding/base64"
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
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"

	solaceconfig "github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/config"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/mocks"
)

// Receiver implements the Receiver for Logs and Traces
type Receiver struct {
	logsConsumer     consumer.Logs
	tracesConsumer   consumer.Traces
	settings         receiver.Settings
	config           *solaceconfig.Config
	logger           *zap.Logger
	wg               sync.WaitGroup
	messagingService interface{} // can be real SDK or mock
	QueueConsumer    interface{} // stores the used QueueConsumer
}

// NewReceiver creates a new Receiver for Logs and Traces
func NewReceiver(
	settings receiver.Settings,
	config *solaceconfig.Config,
	logsConsumer consumer.Logs,
	tracesConsumer consumer.Traces,
	opts ...interface{},
) (*Receiver, error) {
	randNum := rand.Intn(1000000)
	receiver := &Receiver{
		logsConsumer:   logsConsumer,
		tracesConsumer: tracesConsumer,
		settings:       settings,
		config:         config,
		logger:         settings.TelemetrySettings.Logger,
	}
	receiver.logger.Info("NewReceiver instance created",
		zap.Time("created_at", time.Now()),
		zap.String("queue", config.Queue),
		zap.Int("rand", randNum))

	if len(opts) > 0 {
		receiver.messagingService = opts[0]
	}
	return receiver, nil
}

// Start starts the Receiver
func (r *Receiver) Start(ctx context.Context, host component.Host) error {
	r.logger.Info("Starting Solace OTLP receiver",
		zap.String("host", r.config.Host),
		zap.String("queue", r.config.Queue))

	// MessagingService initialize (SDK or Mock)
	if r.messagingService == nil {
		ms, err := messaging.NewMessagingServiceBuilder().
			FromConfigurationProvider(config.ServicePropertyMap{
				config.TransportLayerPropertyHost:                   r.config.Host,
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
		queueConsumer, err := builder.
			WithMessageListener(r.HandleMessage).
			WithClientName("otlp-receiver").
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
		r.logger.Info("Using Mock MessagingService")
		err = ms.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to Solace (mock): %w", err)
		}
		queueConsumerBuilder := ms.CreateQueueConsumerBuilder()
		queueConsumer, err := queueConsumerBuilder.
			WithMessageListener(r.HandleMessage).
			WithClientName("otlp-receiver-mock").
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
		return fmt.Errorf("unsupported messagingService type")
	}

	r.logger.Info("Solace OTLP receiver started successfully!")
	return nil
}

// Shutdown ends the Receiver
func (r *Receiver) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down Solace OTLP receiver")
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
func (r *Receiver) HandleMessage(msg message.InboundMessage) {
	r.logger.Info("HandleMessage called")
	r.wg.Add(1)
	defer r.wg.Done()

	// Try first as base64-encoded OTLP Log to parse
	payloadStr, ok := msg.GetPayloadAsString()
	if !ok {
		r.logger.Error("Failed to get message payload")
		return
	}

	// Try base64-decoding
	payload, err := base64.StdEncoding.DecodeString(payloadStr)
	if err == nil {
		// Try to parse as OTLP Log
		otlpLogs := plogotlp.NewExportRequest()
		if err := otlpLogs.UnmarshalProto(payload); err == nil {
			if err := r.logsConsumer.ConsumeLogs(context.Background(), otlpLogs.Logs()); err != nil {
				r.logger.Error("Failed to consume logs", zap.Error(err))
				return
			}
			acknowledgeMessage(r, msg)
			return
		}

		// Try to parse as OTLP Trace
		otlpTraces := ptraceotlp.NewExportRequest()
		if err := otlpTraces.UnmarshalProto(payload); err == nil {
			if err := r.tracesConsumer.ConsumeTraces(context.Background(), otlpTraces.Traces()); err != nil {
				r.logger.Error("Failed to consume traces", zap.Error(err))
				return
			}
			acknowledgeMessage(r, msg)
			return
		}
	}

	// Try to parse as JSON Log
	var logData struct {
		TimeUnixNano         int64  `json:"time_unix_nano"`
		ObservedTimeUnixNano int64  `json:"observed_time_unix_nano"`
		SeverityNumber       int32  `json:"severity_number"`
		SeverityText         string `json:"severity_text"`
		Body                 string `json:"body"`
		Attributes           []struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		} `json:"attributes"`
		TraceID   string `json:"trace_id"`
		SpanID    string `json:"span_id"`
		EventName string `json:"event_name,omitempty"`
	}

	if err := json.Unmarshal([]byte(payloadStr), &logData); err == nil {
		// Create OTLP log
		otlpLogs := plogotlp.NewExportRequest()
		logs := otlpLogs.Logs()
		resourceLogs := logs.ResourceLogs().AppendEmpty()
		scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
		logRecord := scopeLogs.LogRecords().AppendEmpty()

		// Set log data
		logRecord.SetTimestamp(pcommon.Timestamp(logData.TimeUnixNano))
		logRecord.SetObservedTimestamp(pcommon.Timestamp(logData.ObservedTimeUnixNano))
		logRecord.SetSeverityNumber(plog.SeverityNumber(logData.SeverityNumber))
		logRecord.SetSeverityText(logData.SeverityText)
		logRecord.Body().SetStr(logData.Body)

		// Set attributes
		for _, attr := range logData.Attributes {
			switch v := attr.Value.(type) {
			case string:
				logRecord.Attributes().PutStr(attr.Key, v)
			case float64:
				logRecord.Attributes().PutDouble(attr.Key, v)
			case bool:
				logRecord.Attributes().PutBool(attr.Key, v)
			case int:
				logRecord.Attributes().PutInt(attr.Key, int64(v))
			}
		}

		// Set trace context if available
		if logData.TraceID != "" {
			traceID, err := hexStringToTraceID(logData.TraceID)
			if err == nil {
				logRecord.SetTraceID(traceID)
			}
		}
		if logData.SpanID != "" {
			spanID, err := hexStringToSpanID(logData.SpanID)
			if err == nil {
				logRecord.SetSpanID(spanID)
			}
		}

		if err := r.logsConsumer.ConsumeLogs(context.Background(), logs); err != nil {
			r.logger.Error("Failed to consume logs", zap.Error(err))
			return
		}
		acknowledgeMessage(r, msg)
		return
	}

	// If no Log, try to parse as JSON Trace
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

	if err := json.Unmarshal([]byte(payloadStr), &traceData); err != nil {
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

	if err := r.tracesConsumer.ConsumeTraces(context.Background(), traces); err != nil {
		r.logger.Error("Failed to consume traces", zap.Error(err))
		return
	}
	acknowledgeMessage(r, msg)
}

// Helper functions
func hexStringToTraceID(s string) (pcommon.TraceID, error) {
	var traceID pcommon.TraceID
	_, err := hex.Decode(traceID[:], []byte(s))
	return traceID, err
}

func hexStringToSpanID(s string) (pcommon.SpanID, error) {
	var spanID pcommon.SpanID
	_, err := hex.Decode(spanID[:], []byte(s))
	return spanID, err
}

func getTrustStorePath() string {
	if path := os.Getenv("SESSION_SSL_TRUST_STORE_DIR"); path != "" {
		return path
	}
	return "truststore"
}

func acknowledgeMessage(r *Receiver, msg message.InboundMessage) {
	r.logger.Info("acknowledgeMessage called")
	r.logger.Info("Trying to acknowledge message", zap.String("queueConsumerType", fmt.Sprintf("%T", r.QueueConsumer)))
	if receiver, ok := r.QueueConsumer.(interface {
		Ack(message.InboundMessage) error
	}); ok {
		err := receiver.Ack(msg)
		if err != nil {
			r.logger.Error("Failed to acknowledge message", zap.Error(err))
		} else {
			r.logger.Info("Message acknowledged successfully")
		}
	} else {
		r.logger.Warn("QueueConsumer does not implement Ack interface; message not acknowledged", zap.String("actualType", fmt.Sprintf("%T", r.QueueConsumer)))
	}
}
