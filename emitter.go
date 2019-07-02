package go2p

import "github.com/olebedev/emitter"

type eventEmitter struct {
	emitter *emitter.Emitter
}

func newEventEmitter() *eventEmitter {
	em := new(eventEmitter)
	em.emitter = new(emitter.Emitter)
	em.emitter.Use("*", emitter.Void)
	return em
}

func (em *eventEmitter) EmitAsync(topic string, args ...interface{}) {
	go em.emitter.Emit(topic, args...)
}

func (em *eventEmitter) On(topic string, handler func(args []interface{})) {
	em.emitter.On(topic, func(ev *emitter.Event) {
		handler(ev.Args)
	})
}
