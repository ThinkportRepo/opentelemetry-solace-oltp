package base

import (
	"sync"

	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/config"
)

// Receiver contains common functionality for all receivers
// The config field is of type *config.Config.
type Receiver struct {
	settings         receiver.Settings
	config           *config.Config
	logger           *zap.Logger
	wg               sync.WaitGroup
	messagingService interface{} // can be real SDK or mock
	queueConsumer    interface{} // stores the used QueueConsumer
}

// NewReceiver creates a new base receiver
func NewReceiver(
	settings receiver.Settings,
	config *config.Config,
	opts ...interface{},
) *Receiver {
	receiver := &Receiver{
		settings: settings,
		config:   config,
		logger:   settings.TelemetrySettings.Logger,
	}

	if len(opts) > 0 {
		receiver.messagingService = opts[0]
	}

	return receiver
}

// GetConfig returns the receiver configuration
func (r *Receiver) GetConfig() *config.Config {
	return r.config
}

// GetLogger returns the receiver logger
func (r *Receiver) GetLogger() *zap.Logger {
	return r.logger
}

// GetMessagingService returns the messaging service
func (r *Receiver) GetMessagingService() interface{} {
	return r.messagingService
}

// SetQueueConsumer sets the queue consumer
func (r *Receiver) SetQueueConsumer(consumer interface{}) {
	r.queueConsumer = consumer
}

// GetQueueConsumer returns the queue consumer
func (r *Receiver) GetQueueConsumer() interface{} {
	return r.queueConsumer
}

// AddToWaitGroup adds to the wait group
func (r *Receiver) AddToWaitGroup() {
	r.wg.Add(1)
}

// DoneFromWaitGroup marks the wait group as done
func (r *Receiver) DoneFromWaitGroup() {
	r.wg.Done()
}

// WaitForWaitGroup waits for the wait group to complete
func (r *Receiver) WaitForWaitGroup() {
	r.wg.Wait()
}
