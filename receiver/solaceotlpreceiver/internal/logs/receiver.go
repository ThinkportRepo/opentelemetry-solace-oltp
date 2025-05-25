package logs

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/message"
	basereceiver "github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/receiver"
)

// Receiver handles log processing
type Receiver struct {
	*basereceiver.BaseReceiver
	consumer consumer.Logs
}

// NewReceiver creates a new logs receiver
func NewReceiver(
	settings receiver.Settings,
	config *basereceiver.BaseReceiver,
	consumer consumer.Logs,
) *Receiver {
	return &Receiver{
		BaseReceiver: config,
		consumer:     consumer,
	}
}

// HandleMessage processes incoming log messages
func (r *Receiver) HandleMessage(message message.InboundMessage) {
	r.AddToWaitGroup()
	defer r.DoneFromWaitGroup()

	// Try to parse as base64-encoded OTLP Logs
	if payload, ok := message.GetPayloadAsBytes(); ok {
		// Try to parse as OTLP Log directly first
		otlpLogs := plogotlp.NewExportRequest()
		if err := otlpLogs.UnmarshalProto(payload); err == nil {
			if err := r.consumer.ConsumeLogs(context.Background(), otlpLogs.Logs()); err != nil {
				r.GetLogger().Error("Failed to consume logs", zap.Error(err))
			}
			return
		}

		// Try base64 decoding
		decoded, err := base64.StdEncoding.DecodeString(string(payload))
		if err == nil {
			// Try to parse decoded content as OTLP Log
			if err := otlpLogs.UnmarshalProto(decoded); err == nil {
				if err := r.consumer.ConsumeLogs(context.Background(), otlpLogs.Logs()); err != nil {
					r.GetLogger().Error("Failed to consume logs", zap.Error(err))
				}
				return
			}

			// Try to parse decoded content as JSON
			if logs, err := r.parseJSONLogs(decoded); err == nil {
				if err := r.consumer.ConsumeLogs(context.Background(), logs); err != nil {
					r.GetLogger().Error("Failed to consume logs", zap.Error(err))
				}
				return
			}
		}

		// Try to parse as JSON Logs directly
		if logs, err := r.parseJSONLogs(payload); err == nil {
			if err := r.consumer.ConsumeLogs(context.Background(), logs); err != nil {
				r.GetLogger().Error("Failed to consume logs", zap.Error(err))
			}
			return
		}

		// Log the first few bytes of the payload for debugging
		r.GetLogger().Debug("Failed to parse log message",
			zap.ByteString("payload_start", payload[:min(32, len(payload))]),
			zap.Int("payload_length", len(payload)))
	}
}

// parseJSONLogs attempts to parse JSON logs
func (r *Receiver) parseJSONLogs(payload []byte) (plog.Logs, error) {
	var jsonLogs map[string]interface{}
	if err := json.Unmarshal(payload, &jsonLogs); err != nil {
		return plog.Logs{}, err
	}

	logs := plog.NewLogs()
	// TODO: Implement JSON to OTLP conversion
	return logs, nil
}
