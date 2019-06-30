package go2p

import (
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/olebedev/emitter"
	"github.com/pkg/errors"
)

type Peer struct {
	io         *adapterIO
	send       chan *Message
	middleware middlewares
	emitter    *emitter.Emitter
	metadata   maps.Map
	wg         *sync.WaitGroup
}

func newPeer(adapter Adapter, middleware middlewares) *Peer {
	p := new(Peer)
	p.send = make(chan *Message)
	p.io = newAdapterIO(adapter)
	p.wg = new(sync.WaitGroup)
	p.middleware = middleware
	p.metadata = hashmap.New()
	p.emitter = emitter.New(10)
	p.emitter.Use("*", emitter.Void)

	return p
}

func (p *Peer) start() {
	p.io.emitter.On("disconnect", func(ev *emitter.Event) {
		go p.emitter.Emit("disconnect", p)
	})
	p.io.emitter.On("error", func(ev *emitter.Event) {
		go p.emitter.Emit("error", p, ev.Args[0])
	})

	p.io.start()

	p.wg.Add(1)
	go func(p *Peer) {
		defer func() {
			p.io.adapter.Close()
			p.wg.Done()
		}()

		for {
			select {
			case m := <-p.io.receive:
				p.processPipe(m, Receive)
				continue
			case m := <-p.send:
				p.processPipe(m, Send)
				continue
			case <-p.io.ctx.Done():
				return
			}
		}
	}(p)
}

func (p *Peer) processPipe(m *Message, op PipeOperation) {
	pipe := newPipe(p, p.middleware, op, 0)
	err := pipe.process(m)

	if err == PipeStopProcessing {
		return
	}

	if err != nil {
		p.emitter.Emit("error", p, errors.Wrap(err, "error during process pipe"))
		p.io.cancel()
		return
	}

	if op == Receive {
		p.emitter.Emit("message", p, m)
	} else {
		err := p.io.sendMsg(m)
		if err != nil {
			p.emitter.Emit("error", p, errors.Wrap(err, "error during process pipe"))
			p.io.cancel()
			return
		}
	}

}

func (p *Peer) stop() {
	p.io.cancel()
	p.io.wg.Wait()
	p.wg.Wait()
}

func (p *Peer) Address() string {
	return p.io.adapter.Address()
}

func (p *Peer) Metadata() maps.Map {
	return p.metadata
}
