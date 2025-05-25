package traces

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/base"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/message"
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
func (r *Receiver) HandleMessage(message message.InboundMessage) {
	r.AddToWaitGroup()
	defer r.DoneFromWaitGroup()

	// Try to parse as base64-encoded OTLP Traces
	if payload, ok := message.GetPayloadAsBytes(); ok {
		// Try to parse as OTLP Trace directly first
		otlpTraces := ptraceotlp.NewExportRequest()
		if err := otlpTraces.UnmarshalProto(payload); err == nil {
			if err := r.consumer.ConsumeTraces(context.Background(), otlpTraces.Traces()); err != nil {
				r.GetLogger().Error("Failed to consume traces", zap.Error(err))
			}
			return
		}

		// Try base64 decoding
		decoded, err := base64.StdEncoding.DecodeString(string(payload))
		if err == nil {
			// Try to parse decoded content as OTLP Trace
			if err := otlpTraces.UnmarshalProto(decoded); err == nil {
				if err := r.consumer.ConsumeTraces(context.Background(), otlpTraces.Traces()); err != nil {
					r.GetLogger().Error("Failed to consume traces", zap.Error(err))
				}
				return
			}

			// Try to parse decoded content as JSON
			if traces, err := r.parseJSONTraces(decoded); err == nil {
				if err := r.consumer.ConsumeTraces(context.Background(), traces); err != nil {
					r.GetLogger().Error("Failed to consume traces", zap.Error(err))
				}
				return
			}
		}

		// Try to parse as JSON Traces directly
		if traces, err := r.parseJSONTraces(payload); err == nil {
			if err := r.consumer.ConsumeTraces(context.Background(), traces); err != nil {
				r.GetLogger().Error("Failed to consume traces", zap.Error(err))
			}
			return
		}

		// Log the first few bytes of the payload for debugging
		r.GetLogger().Debug("Failed to parse trace message",
			zap.ByteString("payload_start", payload[:min(32, len(payload))]),
			zap.Int("payload_length", len(payload)))
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseBase64Traces attempts to parse base64-encoded OTLP traces
func (r *Receiver) parseBase64Traces(payload []byte) (ptrace.Traces, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(payload))
	if err != nil {
		return ptrace.Traces{}, err
	}

	otlpTraces := ptraceotlp.NewExportRequest()
	if err := otlpTraces.UnmarshalProto(decoded); err != nil {
		return ptrace.Traces{}, err
	}

	return otlpTraces.Traces(), nil
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
