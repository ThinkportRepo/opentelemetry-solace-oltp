package solaceotlpreceiver

import (
	"bytes"
	"fmt"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"solace.dev/go/messaging/pkg/solace/message"
)

func parseOTLPTraceMessage(msg message.InboundMessage) (ptrace.Traces, error) {
	payload, err := msg.GetPayloadAsBytes()
	if err != nil {
		return ptrace.Traces{}, fmt.Errorf("failed to get message payload: %w", err)
	}

	request := ptraceotlp.NewExportRequest()
	if err := request.UnmarshalProto(bytes.NewReader(payload)); err != nil {
		return ptrace.Traces{}, fmt.Errorf("failed to unmarshal OTLP trace request: %w", err)
	}

	return request.Traces(), nil
}

func parseOTLPLogMessage(msg message.InboundMessage) (plog.Logs, error) {
	payload, err := msg.GetPayloadAsBytes()
	if err != nil {
		return plog.Logs{}, fmt.Errorf("failed to get message payload: %w", err)
	}

	request := plogotlp.NewExportRequest()
	if err := request.UnmarshalProto(bytes.NewReader(payload)); err != nil {
		return plog.Logs{}, fmt.Errorf("failed to unmarshal OTLP log request: %w", err)
	}

	return request.Logs(), nil
}
