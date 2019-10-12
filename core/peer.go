package core

import (
	"sync"
	"sync/atomic"

	"github.com/v-braun/awaiter"
	"github.com/v-braun/go2p/core/utils"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/hashmap"
)

// Peer represents a connection to a remote peer
type Peer struct {
	conn       *connHost
	send       chan *Message
	middleware middlewares
	emitter    *utils.EventEmitter
	metadata   maps.Map
	awaiter    awaiter.Awaiter

	stopping        uint32
	notifyStopOnce  *sync.Once
	notifyErrorOnce *sync.Once
}

func newPeer(conn Conn, middleware middlewares) *Peer {
	p := new(Peer)
	p.send = make(chan *Message, 10)
	p.conn = newConnHost(conn)
	p.awaiter = awaiter.New()
	p.middleware = middleware
	p.metadata = hashmap.New()
	p.emitter = utils.NewEventEmitter()
	p.stopping = 0

	return p
}

func (p *Peer) start() <-chan struct{} {
	done := make(chan struct{})
	p.conn.emitter.Once("disconnect", func() {
		p.stop()
	})
	p.conn.emitter.Once("error", func(err error) {
		go p.emitter.Emit("error", p, err)
		p.stop()
	})

	p.conn.start()

	p.awaiter.Go(func() {
		close(done)
		for {
			select {
			case m := <-p.conn.receive:
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
		p.conn.handleError(err, "processPipe")
		p.stop()
		return
	}

	if op == Receive {
		p.emitter.Emit("message", p, m)
	} else {
		err := p.conn.sendMsg(m)
		if err != nil {
			p.conn.handleError(err, "processPipe")
			p.stop()
			return
		}
	}

}

func (p *Peer) stopInternal() {
	p.conn.Close()
	p.conn.awaiter.Cancel()
	p.awaiter.Cancel()

	p.conn.awaiter.AwaitSync()
	p.awaiter.AwaitSync()
	go p.emitter.Emit("disconnect", p)
}

func (p *Peer) stop() {

	if atomic.LoadUint32(&p.stopping) == 1 {
		return
	}

	atomic.StoreUint32(&p.stopping, 1)
	p.stopInternal()
}

// RemoteAddress returns the remote address of the current peer
func (p *Peer) RemoteAddress() string {
	return p.conn.RemoteAddress()
}

// LocalAddress returns the local address of the current peer
func (p *Peer) LocalAddress() string {
	return p.conn.LocalAddress()
}

// Metadata returns a map of metadata associated to this peer
func (p *Peer) Metadata() maps.Map {
	return p.metadata
}
