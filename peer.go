package go2p

import (
	"github.com/v-braun/awaiter"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/hashmap"
)

// Peer represents a connection to a remote peer
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
	p.send = make(chan *Message, 10)
	p.io = newAdapterIO(adapter)
	p.awaiter = awaiter.New()
	p.middleware = middleware
	p.metadata = hashmap.New()
	p.emitter = newEventEmitter()

	return p
}

func (p *Peer) start() <-chan struct{} {
	done := make(chan struct{})
	p.io.emitter.On("disconnect", func(args []interface{}) {
		p.emitter.EmitAsync("disconnect", p)
	})
	p.io.emitter.On("error", func(args []interface{}) {
		p.emitter.EmitAsync("error", p, args[0])
	})

	p.io.start()

	p.awaiter.Go(func() {
		close(done)
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

	return done
}

func (p *Peer) processPipe(m *Message, op PipeOperation) {
	from := 0
	to := len(p.middleware)
	pos := 0
	if op == Receive {
		pos = to
	}

	pipe := newPipe(p, p.middleware, op, pos, from, to)
	err := pipe.process(m)

	if err == ErrPipeStopProcessing {
		return
	}

	if err != nil {
		p.io.handleError(err, "processPipe")
		p.stopInternal()
		return
	}

	if op == Receive {
		p.emitter.EmitAsync("message", p, m)
	} else {
		err := p.io.sendMsg(m)
		if err != nil {
			p.io.handleError(err, "processPipe")
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

// RemoteAddress returns the remote address of the current peer
func (p *Peer) RemoteAddress() string {
	return p.io.adapter.RemoteAddress()
}

// LocalAddress returns the local address of the current peer
func (p *Peer) LocalAddress() string {
	return p.io.adapter.LocalAddress()
}

// Metadata returns a map of metadata associated to this peer
func (p *Peer) Metadata() maps.Map {
	return p.metadata
}
