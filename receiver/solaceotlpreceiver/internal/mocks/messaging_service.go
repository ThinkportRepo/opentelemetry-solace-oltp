package mocks

import (
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
)

// MessagingService defines the interface for mocking the Solace messaging service
type MessagingService interface {
	Connect() error
	Disconnect() error
	CreateQueueConsumerBuilder() QueueConsumerBuilder
}

// QueueConsumerBuilder defines the interface for mocking the queue consumer builder
type QueueConsumerBuilder interface {
	WithMessageListener(func(message.InboundMessage)) QueueConsumerBuilder
	WithClientName(string) QueueConsumerBuilder
	Build(resource.Queue) (interface{ Start() error }, error)
}
