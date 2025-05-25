package receiver

import (
	"sync"

	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	solaceconfig "github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/config"
)

// BaseReceiver contains common functionality for all receivers
type BaseReceiver struct {
	settings         receiver.Settings
	config           *solaceconfig.Config
	logger           *zap.Logger
	wg               sync.WaitGroup
	messagingService interface{} // can be real SDK or mock
	queueConsumer    interface{} // stores the used QueueConsumer
}

// NewBaseReceiver creates a new base receiver
func NewBaseReceiver(
	settings receiver.Settings,
	config *solaceconfig.Config,
	opts ...interface{},
) *BaseReceiver {
	receiver := &BaseReceiver{
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
func (r *BaseReceiver) GetConfig() *solaceconfig.Config {
	return r.config
}

// GetLogger returns the receiver logger
func (r *BaseReceiver) GetLogger() *zap.Logger {
	return r.logger
}

// GetMessagingService returns the messaging service
func (r *BaseReceiver) GetMessagingService() interface{} {
	return r.messagingService
}

// SetQueueConsumer sets the queue consumer
func (r *BaseReceiver) SetQueueConsumer(consumer interface{}) {
	r.queueConsumer = consumer
}

// GetQueueConsumer returns the queue consumer
func (r *BaseReceiver) GetQueueConsumer() interface{} {
	return r.queueConsumer
}

// AddToWaitGroup adds to the wait group
func (r *BaseReceiver) AddToWaitGroup() {
	r.wg.Add(1)
}

// DoneFromWaitGroup marks the wait group as done
func (r *BaseReceiver) DoneFromWaitGroup() {
	r.wg.Done()
}

// WaitForWaitGroup waits for the wait group to complete
func (r *BaseReceiver) WaitForWaitGroup() {
	r.wg.Wait()
}
