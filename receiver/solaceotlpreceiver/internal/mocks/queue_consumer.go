package mocks

type QueueConsumer interface {
	Start() error
	Terminate(timeout uint) error
}
