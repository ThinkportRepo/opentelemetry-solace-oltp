package receiver

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/util"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.uber.org/zap"
	"solace.dev/go/messaging/pkg/solace/message"
)

// LogsReceiver implements the receiver for logs
type LogsReceiver struct {
	*BaseReceiver
	logsConsumer consumer.Logs
}

// NewLogsReceiver creates a new logs receiver
func NewLogsReceiver(
	base *BaseReceiver,
	logsConsumer consumer.Logs,
) *LogsReceiver {
	return &LogsReceiver{
		BaseReceiver: base,
		logsConsumer: logsConsumer,
	}
}

// HandleMessage processes an incoming message for logs
func (r *LogsReceiver) HandleMessage(msg message.InboundMessage) {
	r.AddToWaitGroup()
	defer r.DoneFromWaitGroup()

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
			}
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
		logs := plog.NewLogs()
		resourceLogs := logs.ResourceLogs().AppendEmpty()
		scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
		logRecord := scopeLogs.LogRecords().AppendEmpty()

		// Set log record data
		logRecord.SetTimestamp(pcommon.Timestamp(logData.TimeUnixNano))
		logRecord.SetObservedTimestamp(pcommon.Timestamp(logData.ObservedTimeUnixNano))
		logRecord.SetSeverityNumber(plog.SeverityNumber(logData.SeverityNumber))
		logRecord.SetSeverityText(logData.SeverityText)
		logRecord.Body().SetStr(logData.Body)

		// Set attributes
		for _, attr := range logData.Attributes {
			logRecord.Attributes().PutStr(attr.Key, attr.Value.(string))
		}

		// Set trace context if available
		if logData.TraceID != "" {
			traceID, err := util.HexStringToTraceID(logData.TraceID)
			if err == nil {
				logRecord.SetTraceID(traceID)
			}
		}
		if logData.SpanID != "" {
			spanID, err := util.HexStringToSpanID(logData.SpanID)
			if err == nil {
				logRecord.SetSpanID(spanID)
			}
		}

		// Consume the logs
		if err := r.logsConsumer.ConsumeLogs(context.Background(), logs); err != nil {
			r.logger.Error("Failed to consume logs", zap.Error(err))
		}
	}
}
