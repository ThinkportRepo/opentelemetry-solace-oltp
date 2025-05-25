package solace

import (
	"time"

	"solace.dev/go/messaging/pkg/solace/message"
)

// QueueConsumer defines the interface for a Queue-Consumer
type QueueConsumer interface {
	// SetMessageListener sets the MessageListener for the consumer
	SetMessageListener(listener MessageListener)
	// Start starts the consumer
	Start() error
	// Terminate terminates the consumer
	Terminate(timeout time.Duration) error
	// ReceiveMessage receives a message with timeout
	ReceiveMessage(timeout time.Duration) (message.InboundMessage, error)
	// Ack confirms a message
	Ack(msg message.InboundMessage) error
}

// PersistentMessageReceiverAdapter is an adapter for the PersistentMessageReceiver
type PersistentMessageReceiverAdapter struct {
	receiver interface {
		Start() error
		Terminate(timeout time.Duration) error
		ReceiveMessage(timeout time.Duration) (message.InboundMessage, error)
		Ack(msg message.InboundMessage) error
	}
	listener MessageListener
}

// NewPersistentMessageReceiverAdapter creates a new adapter
func NewPersistentMessageReceiverAdapter(receiver interface {
	Start() error
	Terminate(timeout time.Duration) error
	ReceiveMessage(timeout time.Duration) (message.InboundMessage, error)
	Ack(msg message.InboundMessage) error
}) *PersistentMessageReceiverAdapter {
	return &PersistentMessageReceiverAdapter{receiver: receiver}
}

// SetMessageListener sets the MessageListener for the consumer
func (a *PersistentMessageReceiverAdapter) SetMessageListener(listener MessageListener) {
	a.listener = listener
}

// Start starts the consumer
func (a *PersistentMessageReceiverAdapter) Start() error {
	return a.receiver.Start()
}

// Terminate terminates the consumer
func (a *PersistentMessageReceiverAdapter) Terminate(timeout time.Duration) error {
	return a.receiver.Terminate(timeout)
}

// ReceiveMessage receives a message with timeout
func (a *PersistentMessageReceiverAdapter) ReceiveMessage(timeout time.Duration) (message.InboundMessage, error) {
	return a.receiver.ReceiveMessage(timeout)
}

// Ack confirms a message
func (a *PersistentMessageReceiverAdapter) Ack(msg message.InboundMessage) error {
	return a.receiver.Ack(msg)
}

// OnMessage is called when a new message is received
func (c *PersistentMessageReceiverAdapter) OnMessage(msg message.InboundMessage) {
	if c.listener != nil {
		c.listener.OnMessage(msg)
	}
}
