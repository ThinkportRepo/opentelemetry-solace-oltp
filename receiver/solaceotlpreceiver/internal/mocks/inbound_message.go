package mocks

type InboundMessage interface {
	GetPayloadAsBytes() ([]byte, error)
}
