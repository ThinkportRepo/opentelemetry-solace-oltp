package mocks

type MessagingService interface {
	Connect() error
	Disconnect() error
	CreateQueueConsumerBuilder() QueueConsumerBuilder
}
