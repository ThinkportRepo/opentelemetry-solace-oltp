package mocks

import (
	"solace.dev/go/messaging/pkg/solace/config"
)

// MessagingServiceBuilder ist ein Interface für den MessagingServiceBuilder
type MessagingServiceBuilder interface {
	FromConfigurationProvider(config.ServicePropertyMap) MessagingServiceBuilder
	Build() (interface{}, error)
}

// MessagingServiceFactory ist ein Interface für die Factory-Methoden
type MessagingServiceFactory interface {
	NewMessagingServiceBuilder() MessagingServiceBuilder
	CreateQueueConsumerBuilder() interface{}
}
