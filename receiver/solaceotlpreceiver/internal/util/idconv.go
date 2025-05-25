package util

import (
	"encoding/hex"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

// HexStringToTraceID converts a hex string to a TraceID
func HexStringToTraceID(s string) (pcommon.TraceID, error) {
	var traceID pcommon.TraceID
	_, err := hex.Decode(traceID[:], []byte(s))
	return traceID, err
}

// HexStringToSpanID converts a hex string to a SpanID
func HexStringToSpanID(s string) (pcommon.SpanID, error) {
	var spanID pcommon.SpanID
	_, err := hex.Decode(spanID[:], []byte(s))
	return spanID, err
}
