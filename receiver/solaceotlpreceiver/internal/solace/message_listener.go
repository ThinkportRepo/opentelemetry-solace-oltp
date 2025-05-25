package solace

import "solace.dev/go/messaging/pkg/solace/message"

// MessageListener definiert die Schnittstelle für die Verarbeitung eingehender Nachrichten
type MessageListener interface {
	// OnMessage wird aufgerufen, wenn eine neue Nachricht empfangen wird
	OnMessage(message message.InboundMessage)
}
