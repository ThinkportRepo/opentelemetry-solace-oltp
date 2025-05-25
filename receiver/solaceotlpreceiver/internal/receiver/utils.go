package receiver

import (
	"encoding/hex"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

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
