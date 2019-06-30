package go2p

import (
	"github.com/pkg/errors"

	"github.com/olebedev/emitter"
)

type NetworkConnectionBuilder struct {
	middlewares []*Middleware
	operators   []PeerOperator
	peerStore   PeerStore
}

func NewNetworkConnection() *NetworkConnectionBuilder {
	b := new(NetworkConnectionBuilder)
	b.peerStore = NewDefaultPeerStore(10)

	return b
}

func (b *NetworkConnectionBuilder) WithMiddleware(name string, impl MiddlewareFunc) *NetworkConnectionBuilder {
	m := NewMiddleware(name, impl)
	b.middlewares = append(b.middlewares, m)
	return b
}

func (b *NetworkConnectionBuilder) WithOperator(op PeerOperator) *NetworkConnectionBuilder {
	b.operators = append(b.operators, op)
	return b
}

func (b *NetworkConnectionBuilder) WithPeerStore(ps PeerStore) *NetworkConnectionBuilder {
	b.peerStore = ps
	return b
}

func (b *NetworkConnectionBuilder) Build() *NetworkConnection {
	nc := new(NetworkConnection)
	nc.middlewares = b.middlewares
	nc.operators = b.operators
	nc.emitter = emitter.New(50)
	nc.emitter.Use("*", emitter.Void)
	nc.peerStore = b.peerStore

	return nc
}

func NewNetworkConnectionTCP(localAddr string) *NetworkConnection {
	op := NewTcpOperator("tcp", localAddr)
	peerStore := NewDefaultPeerStore(10)
	conn := NewNetworkConnection().
		WithOperator(op).
		WithPeerStore(peerStore).
		WithMiddleware(Headers()).
		WithMiddleware(Crypt()).
		WithMiddleware(Log()).
		Build()

	return conn
}

type NetworkConnection struct {
	middlewares []*Middleware
	operators   []PeerOperator
	emitter     *emitter.Emitter
	peerStore   PeerStore
}

func (nc *NetworkConnection) Send(msg *Message, addr string) {
	nc.peerStore.LockPeer(addr, func(peer *Peer) {
		peer.send <- msg
	})
}

func (nc *NetworkConnection) ConnectTo(network string, addr string) {
	for _, op := range nc.operators {
		op.Dial(network, addr)
	}
}

func (nc *NetworkConnection) Start() error {
	nc.peerStore.OnPeerAdd(func(peer *Peer) {
		go nc.emitter.Emit("peer-new", peer)
	})
	nc.peerStore.OnPeerWantRemove(func(peer *Peer) {
		peer.stop()
		nc.peerStore.RemovePeer(peer)
	})

	for _, op := range nc.operators {
		op.OnPeer(func(a Adapter) {
			p := newPeer(a, newMiddlewares(nc.middlewares...))
			err := nc.peerStore.AddPeer(p)
			if err != nil {
				p.emitter.Emit("error", errors.Wrapf(err, "could not add peer: %s", p.Address()))
				return
			}

			p.emitter.On("message", func(ev *emitter.Event) {
				go nc.emitter.Emit("peer-message", ev.Args[0], ev.Args[1])
			})
			p.emitter.On("disconnect", func(ev *emitter.Event) {
				p := ev.Args[0].(*Peer)
				p.stop()
				nc.peerStore.RemovePeer(p)
				nc.emitter.Emit("peer-disconnect", p)
			})
			p.emitter.On("error", func(ev *emitter.Event) {
				p := ev.Args[0].(*Peer)
				p.stop()
				nc.peerStore.RemovePeer(p)
				go nc.emitter.Emit("peer-error", ev.Args[0], ev.Args[1])
			})

			p.start()

			go nc.emitter.Emit("new-peer", p)
		})

		err := op.Start()
		if err != nil {
			return err
		}
	}

	nc.peerStore.Start()

	return nil
}

func (nc *NetworkConnection) OnPeer(handler func(p *Peer)) {
	nc.emitter.On("new-peer", func(ev *emitter.Event) {
		handler(ev.Args[0].(*Peer))
	})
}

func (nc *NetworkConnection) OnMessage(handler func(p *Peer, msg *Message)) {
	nc.emitter.On("peer-message", func(ev *emitter.Event) {
		handler(ev.Args[0].(*Peer), ev.Args[1].(*Message))
	})
}

func (nc *NetworkConnection) OnPeerError(handler func(p *Peer, err error)) {
	nc.emitter.On("peer-error", func(ev *emitter.Event) {
		handler(ev.Args[0].(*Peer), ev.Args[1].(error))
	})
}

func (nc *NetworkConnection) OnPeerDisconnect(handler func(p *Peer)) {
	nc.emitter.On("peer-disconnect", func(ev *emitter.Event) {
		handler(ev.Args[0].(*Peer))
	})
}

func (nc *NetworkConnection) Stop() {
	for _, op := range nc.operators {
		op.Stop()
	}

	nc.peerStore.IteratePeer(func(p *Peer) {
		nc.peerStore.RemovePeer(p)
		p.stop()
	})

	nc.peerStore.Stop()
}
