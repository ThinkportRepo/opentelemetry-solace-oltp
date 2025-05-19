package mocks

import (
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
)

type QueueConsumerBuilder interface {
	WithMessageAutoAcknowledgement() QueueConsumerBuilder
	WithMessageListener(listener func(message.InboundMessage)) QueueConsumerBuilder
	Build(queue resource.Queue) (QueueConsumer, error)
}
