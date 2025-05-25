package solace

import (
	"time"

	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/message/rgmid"
	"solace.dev/go/messaging/pkg/solace/message/sdt"
)

// SolaceInboundMessage wraps a Solace InboundMessage
type SolaceInboundMessage struct {
	msg message.InboundMessage
}

// NewSolaceInboundMessage creates a new SolaceInboundMessage
func NewSolaceInboundMessage(msg message.InboundMessage) message.InboundMessage {
	return &SolaceInboundMessage{msg: msg}
}

// GetPayload returns the message payload
func (m *SolaceInboundMessage) GetPayload() []byte {
	payload, _ := m.msg.GetPayloadAsBytes()
	return payload
}

// GetPayloadAsString returns the message payload as string
func (m *SolaceInboundMessage) GetPayloadAsString() (string, bool) {
	payload, ok := m.msg.GetPayloadAsBytes()
	if !ok || payload == nil {
		return "", false
	}
	return string(payload), true
}

// GetPayloadAsBytes returns the payload as bytes
func (m *SolaceInboundMessage) GetPayloadAsBytes() ([]byte, bool) {
	return m.msg.GetPayloadAsBytes()
}

// GetPayloadAsMap returns the payload as a map
func (m *SolaceInboundMessage) GetPayloadAsMap() (sdt.Map, bool) {
	return m.msg.GetPayloadAsMap()
}

// GetPayloadAsStream returns the payload as a stream
func (m *SolaceInboundMessage) GetPayloadAsStream() (sdt.Stream, bool) {
	return m.msg.GetPayloadAsStream()
}

// Dispose disposes the message
func (m *SolaceInboundMessage) Dispose() {
	m.msg.Dispose()
}

// GetApplicationMessageID returns the application message ID
func (m *SolaceInboundMessage) GetApplicationMessageID() (string, bool) {
	return m.msg.GetApplicationMessageID()
}

// GetApplicationMessageType returns the application message type
func (m *SolaceInboundMessage) GetApplicationMessageType() (string, bool) {
	return m.msg.GetApplicationMessageType()
}

// GetClassOfService returns the Class of Service
func (m *SolaceInboundMessage) GetClassOfService() int {
	return m.msg.GetClassOfService()
}

// GetCorrelationID returns the Correlation ID
func (m *SolaceInboundMessage) GetCorrelationID() (string, bool) {
	return m.msg.GetCorrelationID()
}

// GetDestinationName returns the destination name
func (m *SolaceInboundMessage) GetDestinationName() string {
	return m.msg.GetDestinationName()
}

// GetExpiration returns the expiration date
func (m *SolaceInboundMessage) GetExpiration() time.Time {
	return m.msg.GetExpiration()
}

// GetHTTPContentEncoding returns the HTTP Content Encoding
func (m *SolaceInboundMessage) GetHTTPContentEncoding() (string, bool) {
	return m.msg.GetHTTPContentEncoding()
}

// GetHTTPContentType returns the HTTP Content Type
func (m *SolaceInboundMessage) GetHTTPContentType() (string, bool) {
	return m.msg.GetHTTPContentType()
}

// GetMessageDiscardNotification returns the Message Discard Notification
func (m *SolaceInboundMessage) GetMessageDiscardNotification() message.MessageDiscardNotification {
	return m.msg.GetMessageDiscardNotification()
}

// GetPriority returns the message priority
func (m *SolaceInboundMessage) GetPriority() (int, bool) {
	return m.msg.GetPriority()
}

// GetProperties returns the message properties
func (m *SolaceInboundMessage) GetProperties() sdt.Map {
	return m.msg.GetProperties()
}

// GetProperty returns the message property
func (m *SolaceInboundMessage) GetProperty(name string) (sdt.Data, bool) {
	return m.msg.GetProperty(name)
}

// GetReplicationGroupMessageID returns the ReplicationGroupMessageID of the message
func (m *SolaceInboundMessage) GetReplicationGroupMessageID() (rgmid.ReplicationGroupMessageID, bool) {
	return m.msg.GetReplicationGroupMessageID()
}

// GetSenderID returns the SenderID
func (m *SolaceInboundMessage) GetSenderID() (string, bool) {
	return m.msg.GetSenderID()
}

// GetSenderTimestamp returns the sender timestamp
func (m *SolaceInboundMessage) GetSenderTimestamp() (time.Time, bool) {
	return m.msg.GetSenderTimestamp()
}

// GetSequenceNumber returns the sequence number
func (m *SolaceInboundMessage) GetSequenceNumber() (int64, bool) {
	return m.msg.GetSequenceNumber()
}

// HasProperty checks if the message has a property with the given name
func (m *SolaceInboundMessage) HasProperty(name string) bool {
	return m.msg.HasProperty(name)
}

// GetTimeStamp returns the message timestamp
func (m *SolaceInboundMessage) GetTimeStamp() (time.Time, bool) {
	return m.msg.GetTimeStamp()
}

// IsDisposed checks if the message has already been disposed
func (m *SolaceInboundMessage) IsDisposed() bool {
	return m.msg.IsDisposed()
}

// IsRedelivered checks if the message has been redelivered
func (m *SolaceInboundMessage) IsRedelivered() bool {
	return m.msg.IsRedelivered()
}

// String returns a string representation of the message
func (m *SolaceInboundMessage) String() string {
	return "SolaceInboundMessage"
}
