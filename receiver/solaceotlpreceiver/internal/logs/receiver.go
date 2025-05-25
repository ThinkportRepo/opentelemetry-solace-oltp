package logs

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/base"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/interfaces"
)

// Receiver handles log processing
type Receiver struct {
	*base.Receiver
	consumer consumer.Logs
}

// NewReceiver creates a new logs receiver
func NewReceiver(
	settings receiver.Settings,
	config *base.Receiver,
	consumer consumer.Logs,
) *Receiver {
	return &Receiver{
		Receiver: config,
		consumer: consumer,
	}
}

// HandleMessage processes incoming log messages
func (r *Receiver) HandleMessage(message interfaces.InboundMessage) {
	r.AddToWaitGroup()
	defer r.DoneFromWaitGroup()

	// Try to parse as base64-encoded OTLP Logs
	if logs, err := r.parseBase64Logs(message.GetPayload()); err == nil {
		if err := r.consumer.ConsumeLogs(context.Background(), logs); err != nil {
			r.GetLogger().Error("Failed to consume logs", zap.Error(err))
		}
		return
	}

	// Try to parse as JSON Logs
	if logs, err := r.parseJSONLogs(message.GetPayload()); err == nil {
		if err := r.consumer.ConsumeLogs(context.Background(), logs); err != nil {
			r.GetLogger().Error("Failed to consume logs", zap.Error(err))
		}
		return
	}

	r.GetLogger().Warn("Failed to parse log message", zap.ByteString("payload", message.GetPayload()))
}

// parseBase64Logs attempts to parse base64-encoded OTLP logs
func (r *Receiver) parseBase64Logs(payload []byte) (plog.Logs, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(payload))
	if err != nil {
		return plog.Logs{}, err
	}

	logs := plog.NewLogs()
	if err := logs.UnmarshalProto(decoded); err != nil {
		return plog.Logs{}, err
	}

	return logs, nil
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
