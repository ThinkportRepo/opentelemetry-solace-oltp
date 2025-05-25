package solace

import (
	"time"

	"solace.dev/go/messaging/pkg/solace/message"
)

// QueueConsumer definiert die Schnittstelle für einen Queue-Consumer
type QueueConsumer interface {
	// SetMessageListener setzt den MessageListener für den Consumer
	SetMessageListener(listener MessageListener)
	// Start startet den Consumer
	Start() error
	// Terminate beendet den Consumer
	Terminate(timeout time.Duration) error
	// ReceiveMessage empfängt eine Nachricht mit Timeout
	ReceiveMessage(timeout time.Duration) (message.InboundMessage, error)
	// Ack bestätigt eine Nachricht
	Ack(msg message.InboundMessage) error
}

// PersistentMessageReceiverAdapter ist ein Adapter für den PersistentMessageReceiver
type PersistentMessageReceiverAdapter struct {
	receiver interface {
		Start() error
		Terminate(timeout time.Duration) error
		ReceiveMessage(timeout time.Duration) (message.InboundMessage, error)
		Ack(msg message.InboundMessage) error
	}
	listener MessageListener
}

// NewPersistentMessageReceiverAdapter erstellt einen neuen Adapter
func NewPersistentMessageReceiverAdapter(receiver interface {
	Start() error
	Terminate(timeout time.Duration) error
	ReceiveMessage(timeout time.Duration) (message.InboundMessage, error)
	Ack(msg message.InboundMessage) error
}) *PersistentMessageReceiverAdapter {
	return &PersistentMessageReceiverAdapter{receiver: receiver}
}

// SetMessageListener setzt den MessageListener für den Consumer
func (a *PersistentMessageReceiverAdapter) SetMessageListener(listener MessageListener) {
	a.listener = listener
}

// Start startet den Consumer
func (a *PersistentMessageReceiverAdapter) Start() error {
	return a.receiver.Start()
}

// Terminate beendet den Consumer
func (a *PersistentMessageReceiverAdapter) Terminate(timeout time.Duration) error {
	return a.receiver.Terminate(timeout)
}

// ReceiveMessage empfängt eine Nachricht mit Timeout
func (a *PersistentMessageReceiverAdapter) ReceiveMessage(timeout time.Duration) (message.InboundMessage, error) {
	return a.receiver.ReceiveMessage(timeout)
}

// Ack bestätigt eine Nachricht
func (a *PersistentMessageReceiverAdapter) Ack(msg message.InboundMessage) error {
	return a.receiver.Ack(msg)
}

// OnMessage wird aufgerufen, wenn eine neue Nachricht empfangen wird
func (c *PersistentMessageReceiverAdapter) OnMessage(msg message.InboundMessage) {
	if c.listener != nil {
		c.listener.OnMessage(msg)
	}
}
