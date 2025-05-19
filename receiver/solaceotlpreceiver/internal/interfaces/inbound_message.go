package interfaces

import (
	"time"

	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/message/rgmid"
	"solace.dev/go/messaging/pkg/solace/message/sdt"
)

// InboundMessage represents a message received from Solace
type InboundMessage interface {
	// Dispose releases any resources associated with the message
	Dispose()
	// GetProperties returns a map of user properties
	GetProperties() sdt.Map
	// GetProperty returns the user property at the given key
	GetProperty(key string) (value sdt.Data, ok bool)
	// GetPayloadAsBytes returns the message payload as bytes
	GetPayloadAsBytes() ([]byte, bool)
	// GetPayloadAsString returns the message payload as string
	GetPayloadAsString() (string, bool)
	// GetPayloadAsMap returns the message payload as map
	GetPayloadAsMap() (sdt.Map, bool)
	// GetPayloadAsStream returns the message payload as stream
	GetPayloadAsStream() (sdt.Stream, bool)
	// GetApplicationMessageID returns the application message ID
	GetApplicationMessageID() (string, bool)
	// GetApplicationMessageType returns the application message type
	GetApplicationMessageType() (string, bool)
	// GetCorrelationID returns the correlation ID
	GetCorrelationID() (string, bool)
	// GetDestination returns the destination
	GetDestination() (string, bool)
	// GetDestinationName returns the destination name
	GetDestinationName() string
	// GetExpiration returns the expiration time
	GetExpiration() time.Time
	// GetPriority returns the priority
	GetPriority() (int, bool)
	// GetRedelivered returns whether the message was redelivered
	GetRedelivered() bool
	// GetReplyTo returns the reply-to destination
	GetReplyTo() (string, bool)
	// GetTimeToLive returns the time to live
	GetTimeToLive() (int64, bool)
	// GetUserPropertyMap returns the user property map
	GetUserPropertyMap() (map[string]interface{}, bool)
	// IsBinary returns whether the message is binary
	IsBinary() bool
	// IsText returns whether the message is text
	IsText() bool
	// GetClassOfService returns the class of service
	GetClassOfService() int
	// GetHTTPContentEncoding returns the HTTP content encoding
	GetHTTPContentEncoding() (string, bool)
	// GetHTTPContentType returns the HTTP content type
	GetHTTPContentType() (string, bool)
	// GetMessageDiscardNotification returns the discard notification
	GetMessageDiscardNotification() message.MessageDiscardNotification
	// GetReplicationGroupMessageID returns the replication group message ID
	GetReplicationGroupMessageID() (rgmid.ReplicationGroupMessageID, bool)
	// GetSenderID returns the sender ID
	GetSenderID() (string, bool)
	// GetSenderTimestamp returns the sender timestamp
	GetSenderTimestamp() (time.Time, bool)
	// GetSequenceNumber returns the sequence number
	GetSequenceNumber() (int64, bool)
	// GetTimeStamp returns the timestamp
	GetTimeStamp() (time.Time, bool)
	// HasProperty returns whether the message has a property with the given name
	HasProperty(name string) bool
	// IsDisposed returns whether the message is disposed
	IsDisposed() bool
	// IsRedelivered returns whether the message was redelivered
	IsRedelivered() bool
	String() string
}
