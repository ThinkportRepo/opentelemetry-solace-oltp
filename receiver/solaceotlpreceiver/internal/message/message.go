package message

import "solace.dev/go/messaging/pkg/solace/message"

// InboundMessage ist ein Alias für die Solace InboundMessage-Schnittstelle
type InboundMessage = message.InboundMessage

// MessageListener definiert die Schnittstelle für die Verarbeitung eingehender Nachrichten
type MessageListener interface {
	// OnMessage wird aufgerufen, wenn eine neue Nachricht empfangen wird
	OnMessage(message InboundMessage)
}
