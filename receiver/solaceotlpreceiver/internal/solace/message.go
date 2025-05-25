package solace

import (
	"time"

	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/message/rgmid"
	"solace.dev/go/messaging/pkg/solace/message/sdt"
)

// Message represents a message received from Solace
type Message struct {
	payload []byte
}

// NewMessage creates a new Message
func NewMessage(payload []byte) message.InboundMessage {
	return &Message{payload: payload}
}

// GetPayload gibt die Payload der Nachricht zurück
func (m *Message) GetPayload() []byte {
	payload, _ := m.GetPayloadAsBytes()
	return payload
}

// GetPayloadAsString returns the message payload as string
func (m *Message) GetPayloadAsString() (string, bool) {
	if m.payload == nil {
		return "", false
	}
	return string(m.payload), true
}

// GetPayloadAsBytes gibt die Payload als Bytes zurück
func (m *Message) GetPayloadAsBytes() ([]byte, bool) {
	return m.payload, true
}

// GetPayloadAsMap gibt die Payload als Map zurück
func (m *Message) GetPayloadAsMap() (sdt.Map, bool) {
	return nil, false
}

// GetPayloadAsStream gibt die Payload als Stream zurück
func (m *Message) GetPayloadAsStream() (sdt.Stream, bool) {
	return nil, false
}

// Dispose releases any resources associated with the message
func (m *Message) Dispose() {
	// Nothing to dispose
}

// GetApplicationMessageID returns the application message ID
func (m *Message) GetApplicationMessageID() (string, bool) {
	return "", false
}

// GetApplicationMessageType returns the application message type
func (m *Message) GetApplicationMessageType() (string, bool) {
	return "", false
}

// GetClassOfService gibt die Class of Service zurück
func (m *Message) GetClassOfService() int {
	return 0
}

// GetCorrelationID gibt die Correlation ID zurück
func (m *Message) GetCorrelationID() (string, bool) {
	return "", false
}

// GetDestinationName gibt den Zielnamen zurück
func (m *Message) GetDestinationName() string {
	return ""
}

// GetExpiration gibt das Ablaufdatum zurück
func (m *Message) GetExpiration() time.Time {
	return time.Time{}
}

// GetHTTPContentEncoding gibt die HTTP-Content-Encoding zurück
func (m *Message) GetHTTPContentEncoding() (string, bool) {
	return "", false
}

// GetHTTPContentType gibt den HTTP-Content-Type zurück
func (m *Message) GetHTTPContentType() (string, bool) {
	return "", false
}

// GetMessageDiscardNotification gibt die Message-Discard-Notification zurück
func (m *Message) GetMessageDiscardNotification() message.MessageDiscardNotification {
	return nil
}

// GetPriority gibt die Priorität der Nachricht zurück
func (m *Message) GetPriority() (int, bool) {
	return 0, false
}

// GetProperties gibt die Properties der Nachricht zurück
func (m *Message) GetProperties() sdt.Map {
	return nil
}

// GetProperty gibt die Property der Nachricht zurück
func (m *Message) GetProperty(name string) (sdt.Data, bool) {
	return nil, false
}

// GetReplicationGroupMessageID gibt die ReplicationGroupMessageID der Nachricht zurück
func (m *Message) GetReplicationGroupMessageID() (rgmid.ReplicationGroupMessageID, bool) {
	return nil, false
}

// GetSenderID gibt die SenderID der Nachricht zurück
func (m *Message) GetSenderID() (string, bool) {
	return "", false
}

// GetSenderTimestamp gibt den Sender-Timestamp der Nachricht zurück
func (m *Message) GetSenderTimestamp() (time.Time, bool) {
	return time.Time{}, false
}

// GetSequenceNumber gibt die Sequence Number der Nachricht zurück
func (m *Message) GetSequenceNumber() (int64, bool) {
	return 0, false
}

// HasProperty prüft, ob die Nachricht eine Property mit dem gegebenen Namen hat
func (m *Message) HasProperty(name string) bool {
	return false
}

// GetTimeStamp gibt den Zeitstempel der Nachricht zurück
func (m *Message) GetTimeStamp() (time.Time, bool) {
	return time.Time{}, false
}

// IsDisposed prüft, ob die Nachricht bereits entsorgt wurde
func (m *Message) IsDisposed() bool {
	return false
}

// IsRedelivered prüft, ob die Nachricht erneut zugestellt wurde
func (m *Message) IsRedelivered() bool {
	return false
}

// String gibt eine String-Repräsentation der Nachricht zurück
func (m *Message) String() string {
	return "Message"
}
