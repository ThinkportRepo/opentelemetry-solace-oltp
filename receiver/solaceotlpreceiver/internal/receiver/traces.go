package receiver

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
	"solace.dev/go/messaging/pkg/solace/message"
)

// TracesReceiver implements the receiver for traces
type TracesReceiver struct {
	*BaseReceiver
	tracesConsumer consumer.Traces
}

// NewTracesReceiver creates a new traces receiver
func NewTracesReceiver(
	base *BaseReceiver,
	tracesConsumer consumer.Traces,
) *TracesReceiver {
	return &TracesReceiver{
		BaseReceiver:   base,
		tracesConsumer: tracesConsumer,
	}
}

// HandleMessage processes an incoming message for traces
func (r *TracesReceiver) HandleMessage(msg message.InboundMessage) {
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
		// Try to parse as OTLP Trace
		otlpTraces := ptraceotlp.NewExportRequest()
		if err := otlpTraces.UnmarshalProto(payload); err == nil {
			if err := r.tracesConsumer.ConsumeTraces(context.Background(), otlpTraces.Traces()); err != nil {
				r.logger.Error("Failed to consume traces", zap.Error(err))
			}
			return
		}
	}

	// Try to parse as JSON Trace
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

	if err := json.Unmarshal([]byte(payloadStr), &traceData); err == nil {
		// Create OTLP trace
		traces := ptrace.NewTraces()
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

		// Consume the traces
		if err := r.tracesConsumer.ConsumeTraces(context.Background(), traces); err != nil {
			r.logger.Error("Failed to consume traces", zap.Error(err))
		}
	}
}
