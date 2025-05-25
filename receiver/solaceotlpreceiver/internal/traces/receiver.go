package traces

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/base"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/interfaces"
)

// Receiver handles trace processing
type Receiver struct {
	*base.Receiver
	consumer consumer.Traces
}

// NewReceiver creates a new traces receiver
func NewReceiver(
	settings receiver.Settings,
	config *base.Receiver,
	consumer consumer.Traces,
) *Receiver {
	return &Receiver{
		Receiver: config,
		consumer: consumer,
	}
}

// HandleMessage processes incoming trace messages
func (r *Receiver) HandleMessage(message interfaces.InboundMessage) {
	r.AddToWaitGroup()
	defer r.DoneFromWaitGroup()

	// Try to parse as base64-encoded OTLP Traces
	if traces, err := r.parseBase64Traces(message.GetPayload()); err == nil {
		if err := r.consumer.ConsumeTraces(context.Background(), traces); err != nil {
			r.GetLogger().Error("Failed to consume traces", zap.Error(err))
		}
		return
	}

	// Try to parse as JSON Traces
	if traces, err := r.parseJSONTraces(message.GetPayload()); err == nil {
		if err := r.consumer.ConsumeTraces(context.Background(), traces); err != nil {
			r.GetLogger().Error("Failed to consume traces", zap.Error(err))
		}
		return
	}

	r.GetLogger().Warn("Failed to parse trace message", zap.ByteString("payload", message.GetPayload()))
}

// parseBase64Traces attempts to parse base64-encoded OTLP traces
func (r *Receiver) parseBase64Traces(payload []byte) (ptrace.Traces, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(payload))
	if err != nil {
		return ptrace.Traces{}, err
	}

	traces := ptrace.NewTraces()
	if err := traces.UnmarshalProto(decoded); err != nil {
		return ptrace.Traces{}, err
	}

	return traces, nil
}

// parseJSONTraces attempts to parse JSON traces
func (r *Receiver) parseJSONTraces(payload []byte) (ptrace.Traces, error) {
	var jsonTraces map[string]interface{}
	if err := json.Unmarshal(payload, &jsonTraces); err != nil {
		return ptrace.Traces{}, err
	}

	traces := ptrace.NewTraces()
	// TODO: Implement JSON to OTLP conversion
	return traces, nil
}
