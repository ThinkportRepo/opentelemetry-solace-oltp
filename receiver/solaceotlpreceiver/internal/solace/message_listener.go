package solace

import "solace.dev/go/messaging/pkg/solace/message"

// MessageListener defines the interface for processing incoming messages
type MessageListener interface {
	// OnMessage is called when a new message is received
	OnMessage(message message.InboundMessage)
}
