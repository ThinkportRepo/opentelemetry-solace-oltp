package interfaces

// InboundMessage represents a message received from Solace
// Only the methods actually used in the receiver code are included.
type InboundMessage interface {
	// GetPayload returns the message payload
	GetPayload() []byte
	// GetPayloadAsString returns the message payload as string
	GetPayloadAsString() (string, bool)
	// Dispose releases any resources associated with the message
	Dispose()
}
