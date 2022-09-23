package mocks

import (
	"sync"

	"github.com/streadway/amqp"
)

// RabbitClientMock -
type RabbitClientMock struct {
	mut    sync.Mutex
	events map[string]amqp.Publishing
}

// NewRabbitClientMock -
func NewRabbitClientMock() *RabbitClientMock {
	return &RabbitClientMock{
		events: make(map[string]amqp.Publishing),
	}
}

// Publish -
func (rc *RabbitClientMock) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	rc.mut.Lock()
	defer rc.mut.Unlock()

	rc.events[exchange] = msg

	return nil
}

// ExchangeDeclare -
func (rc *RabbitClientMock) ExchangeDeclare(name, kind string) error {
	return nil
}

// ConnErrChan -
func (rc *RabbitClientMock) ConnErrChan() chan *amqp.Error {
	return make(chan *amqp.Error)
}

// CloseErrChan -
func (rc *RabbitClientMock) CloseErrChan() chan *amqp.Error {
	return make(chan *amqp.Error)
}

// Reconnect -
func (rc *RabbitClientMock) Reconnect() {
}

// ReopenChannel -
func (rc *RabbitClientMock) ReopenChannel() {
}

// GetEntries -
func (rc *RabbitClientMock) GetEntries() map[string]amqp.Publishing {
	rc.mut.Lock()
	defer rc.mut.Unlock()

	return rc.events
}

// Close -
func (rc *RabbitClientMock) Close() {
}

// IsInterfaceNil -
func (rc *RabbitClientMock) IsInterfaceNil() bool {
	return rc == nil
}
