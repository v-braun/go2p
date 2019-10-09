package utils

import (
	"github.com/chuckpreslar/emission"
)

type EventEmitter struct {
	emitter *emission.Emitter
}

func NewEventEmitter() *EventEmitter {
	em := new(EventEmitter)
	em.emitter = emission.NewEmitter()

	return em
}

func (em *EventEmitter) Emit(topic string, args ...interface{}) {
	em.emitter.Emit(topic, args...)
}

func (em *EventEmitter) On(topic string, handler interface{}) {
	em.emitter.On(topic, handler)
}
