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

// GetPayload gibt die Payload der Nachricht zurück
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

// GetPayloadAsBytes gibt die Payload als Bytes zurück
func (m *SolaceInboundMessage) GetPayloadAsBytes() ([]byte, bool) {
	return m.msg.GetPayloadAsBytes()
}

// GetPayloadAsMap gibt die Payload als Map zurück
func (m *SolaceInboundMessage) GetPayloadAsMap() (sdt.Map, bool) {
	return nil, false
}

// GetPayloadAsStream gibt die Payload als Stream zurück
func (m *SolaceInboundMessage) GetPayloadAsStream() (sdt.Stream, bool) {
	return nil, false
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

// GetClassOfService gibt die Class of Service zurück
func (m *SolaceInboundMessage) GetClassOfService() int {
	return 0
}

// GetCorrelationID gibt die Correlation ID zurück
func (m *SolaceInboundMessage) GetCorrelationID() (string, bool) {
	return "", false
}

// GetDestinationName gibt den Zielnamen zurück
func (m *SolaceInboundMessage) GetDestinationName() string {
	return ""
}

// GetExpiration gibt das Ablaufdatum zurück
func (m *SolaceInboundMessage) GetExpiration() time.Time {
	return time.Time{}
}

// GetHTTPContentEncoding gibt die HTTP-Content-Encoding zurück
func (m *SolaceInboundMessage) GetHTTPContentEncoding() (string, bool) {
	return "", false
}

// GetHTTPContentType gibt den HTTP-Content-Type zurück
func (m *SolaceInboundMessage) GetHTTPContentType() (string, bool) {
	return "", false
}

// GetMessageDiscardNotification gibt die Message-Discard-Notification zurück
func (m *SolaceInboundMessage) GetMessageDiscardNotification() message.MessageDiscardNotification {
	return nil
}

// GetPriority gibt die Priorität der Nachricht zurück
func (m *SolaceInboundMessage) GetPriority() (int, bool) {
	return m.msg.GetPriority()
}

// GetProperties gibt die Properties der Nachricht zurück
func (m *SolaceInboundMessage) GetProperties() sdt.Map {
	return nil
}

// GetProperty gibt die Property der Nachricht zurück
func (m *SolaceInboundMessage) GetProperty(name string) (sdt.Data, bool) {
	return m.msg.GetProperty(name)
}

// GetReplicationGroupMessageID gibt die ReplicationGroupMessageID der Nachricht zurück
func (m *SolaceInboundMessage) GetReplicationGroupMessageID() (rgmid.ReplicationGroupMessageID, bool) {
	return m.msg.GetReplicationGroupMessageID()
}

// GetSenderID gibt die SenderID der Nachricht zurück
func (m *SolaceInboundMessage) GetSenderID() (string, bool) {
	return "", false
}

// GetSenderTimestamp gibt den Sender-Timestamp der Nachricht zurück
func (m *SolaceInboundMessage) GetSenderTimestamp() (time.Time, bool) {
	return time.Time{}, false
}

// GetSequenceNumber gibt die Sequence Number der Nachricht zurück
func (m *SolaceInboundMessage) GetSequenceNumber() (int64, bool) {
	return 0, false
}

// GetTimeStamp gibt den Zeitstempel der Nachricht zurück
func (m *SolaceInboundMessage) GetTimeStamp() (time.Time, bool) {
	return time.Time{}, false
}

// HasProperty prüft, ob die Nachricht eine Property mit dem gegebenen Namen hat
func (m *SolaceInboundMessage) HasProperty(name string) bool {
	return false
}

// IsDisposed prüft, ob die Nachricht bereits entsorgt wurde
func (m *SolaceInboundMessage) IsDisposed() bool {
	return false
}

// IsRedelivered prüft, ob die Nachricht erneut zugestellt wurde
func (m *SolaceInboundMessage) IsRedelivered() bool {
	return false
}

// String gibt eine String-Repräsentation der Nachricht zurück
func (m *SolaceInboundMessage) String() string {
	return "SolaceInboundMessage"
}
