package go2p

import (
	"github.com/v-braun/awaiter"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/pkg/errors"
)

type Peer struct {
	io         *adapterIO
	send       chan *Message
	middleware middlewares
	emitter    *eventEmitter
	metadata   maps.Map
	awaiter    awaiter.Awaiter
}

func newPeer(adapter Adapter, middleware middlewares) *Peer {
	p := new(Peer)
	p.send = make(chan *Message)
	p.io = newAdapterIO(adapter)
	p.awaiter = awaiter.New()
	p.middleware = middleware
	p.metadata = hashmap.New()
	p.emitter = newEventEmitter()

	return p
}

func (p *Peer) start() {
	p.io.emitter.On("disconnect", func(args []interface{}) {
		p.emitter.EmitAsync("disconnect", p)
	})
	p.io.emitter.On("error", func(args []interface{}) {
		p.emitter.EmitAsync("error", p, args[0])
	})

	p.io.start()

	p.awaiter.Go(func() {
		for {
			select {
			case m := <-p.io.receive:
				p.processPipe(m, Receive)
				continue
			case m := <-p.send:
				p.processPipe(m, Send)
				continue
			case <-p.awaiter.CancelRequested():
				return
			}
		}
	})
}

func (p *Peer) processPipe(m *Message, op PipeOperation) {
	pipe := newPipe(p, p.middleware, op, 0)
	err := pipe.process(m)

	if err == PipeStopProcessing {
		return
	}

	if err != nil {
		p.emitter.EmitAsync("error", p, errors.Wrap(err, "error during process pipe"))
		p.stopInternal()
		return
	}

	if op == Receive {
		p.emitter.EmitAsync("message", p, m)
	} else {
		err := p.io.sendMsg(m)
		if err != nil {
			p.emitter.EmitAsync("error", p, errors.Wrap(err, "error during process pipe"))
			p.stopInternal()
			return
		}
	}

}

func (p *Peer) stopInternal() {
	p.io.adapter.Close()
	p.io.awaiter.Cancel()
	p.awaiter.Cancel()
}

func (p *Peer) stop() {
	p.stopInternal()
	p.io.awaiter.AwaitSync()
	p.awaiter.AwaitSync()
}

func (p *Peer) Address() string {
	return p.io.adapter.Address()
}

func (p *Peer) Metadata() maps.Map {
	return p.metadata
}
