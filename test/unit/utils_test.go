package receiver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/receiver"
)

func TestHexStringToTraceID_Valid(t *testing.T) {
	hexStr := "00112233445566778899aabbccddeeff"
	traceID, err := receiver.HexStringToTraceID(hexStr)
	assert.NoError(t, err)
	assert.Equal(t, hexStr, traceID.String())
}

func TestHexStringToTraceID_Invalid(t *testing.T) {
	hexStr := "invalidhex"
	_, err := receiver.HexStringToTraceID(hexStr)
	assert.Error(t, err)
}

func TestHexStringToSpanID_Valid(t *testing.T) {
	hexStr := "0011223344556677"
	spanID, err := receiver.HexStringToSpanID(hexStr)
	assert.NoError(t, err)
	assert.Equal(t, hexStr, spanID.String())
}

func TestHexStringToSpanID_Invalid(t *testing.T) {
	hexStr := "invalidhex"
	_, err := receiver.HexStringToSpanID(hexStr)
	assert.Error(t, err)
}
