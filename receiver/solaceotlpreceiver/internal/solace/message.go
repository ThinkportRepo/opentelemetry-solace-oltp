package solace

import (
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/interfaces"
	"solace.dev/go/messaging/pkg/solace/message"
)

// InboundMessage represents a message received from Solace
type InboundMessage interface {
	GetPayload() []byte
	GetProperties() map[string]interface{}
	GetDestinationName() string
	GetMessageID() string
	GetCorrelationID() string
	GetApplicationMessageID() string
	GetApplicationMessageType() string
	GetApplicationMessageProperties() map[string]interface{}
	GetApplicationMessageUserPropertyMap() map[string]interface{}
	GetApplicationMessageUserPropertyArray() []interface{}
	GetApplicationMessageUserPropertyString() string
	GetApplicationMessageUserPropertyInt() int64
	GetApplicationMessageUserPropertyFloat() float64
	GetApplicationMessageUserPropertyBool() bool
	GetApplicationMessageUserPropertyBytes() []byte
	GetApplicationMessageUserPropertyTime() int64
	GetApplicationMessageUserPropertyDuration() int64
	GetApplicationMessageUserPropertyMapString() map[string]string
	GetApplicationMessageUserPropertyMapInt() map[string]int64
	GetApplicationMessageUserPropertyMapFloat() map[string]float64
	GetApplicationMessageUserPropertyMapBool() map[string]bool
	GetApplicationMessageUserPropertyMapBytes() map[string][]byte
	GetApplicationMessageUserPropertyMapTime() map[string]int64
	GetApplicationMessageUserPropertyMapDuration() map[string]int64
	GetApplicationMessageUserPropertyArrayString() []string
	GetApplicationMessageUserPropertyArrayInt() []int64
	GetApplicationMessageUserPropertyArrayFloat() []float64
	GetApplicationMessageUserPropertyArrayBool() []bool
	GetApplicationMessageUserPropertyArrayBytes() [][]byte
	GetApplicationMessageUserPropertyArrayTime() []int64
	GetApplicationMessageUserPropertyArrayDuration() []int64
}

// Message represents a Solace message
type Message struct {
	msg message.InboundMessage
}

// NewMessage creates a new Message from a Solace inbound message
func NewMessage(msg message.InboundMessage) interfaces.InboundMessage {
	return &Message{msg: msg}
}

// GetPayload returns the message payload
func (m *Message) GetPayload() []byte {
	payload, _ := m.msg.GetPayloadAsBytes()
	return payload
}

// GetPayloadAsString returns the message payload as string
func (m *Message) GetPayloadAsString() (string, bool) {
	return m.msg.GetPayloadAsString()
}

// GetProperties returns the message properties
func (m *Message) GetProperties() map[string]interface{} {
	props := make(map[string]interface{})
	for k, v := range m.msg.GetProperties() {
		props[k] = v
	}
	return props
}

// GetDestinationName returns the message destination name
func (m *Message) GetDestinationName() string {
	return m.msg.GetDestinationName()
}

// Dispose releases any resources associated with the message
func (m *Message) Dispose() {
	m.msg.Dispose()
}
