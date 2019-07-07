package go2p

import (
	"fmt"

	"github.com/pkg/errors"
)

// NetworkConnectionBuilder provides a fluent interface to
// create a NetworkConnection
type NetworkConnectionBuilder struct {
	middlewares []*Middleware
	operators   []PeerOperator
	peerStore   PeerStore
}

// NewNetworkConnection creates a new NetworkBuilder instance to setup a new NetworkConnection
func NewNetworkConnection() *NetworkConnectionBuilder {
	b := new(NetworkConnectionBuilder)
	b.peerStore = NewDefaultPeerStore(10)

	return b
}

// WithMiddleware attach a new Middleware to the NetworkConnection setup
func (b *NetworkConnectionBuilder) WithMiddleware(name string, impl MiddlewareFunc) *NetworkConnectionBuilder {
	m := NewMiddleware(name, impl)
	b.middlewares = append(b.middlewares, m)
	return b
}

// WithOperator attach a new PeerOperator to the NetworkConnection setup
func (b *NetworkConnectionBuilder) WithOperator(op PeerOperator) *NetworkConnectionBuilder {
	b.operators = append(b.operators, op)
	return b
}

// WithPeerStore attach a new PeerStore to the NetworkConnection
func (b *NetworkConnectionBuilder) WithPeerStore(ps PeerStore) *NetworkConnectionBuilder {
	b.peerStore = ps
	return b
}

// Build finalize the NetworkConnection setup and creates the new instance
func (b *NetworkConnectionBuilder) Build() *NetworkConnection {
	nc := new(NetworkConnection)
	nc.middlewares = newMiddlewares(b.middlewares...)
	nc.operators = b.operators
	nc.emitter = newEventEmitter()
	nc.peerStore = b.peerStore

	return nc
}

/*
NewNetworkConnectionTCP provides a full configured TCP based network
It use the _DefaultMiddleware_ a TCP based operator and the following middleware:

Routes, Headers, Crypt, Log


*/
func NewNetworkConnectionTCP(localAddr string, routes RoutingTable) *NetworkConnection {
	op := NewTcpOperator("tcp", localAddr)
	peerStore := NewDefaultPeerStore(10)

	conn := NewNetworkConnection().
		WithOperator(op).
		WithPeerStore(peerStore).
		WithMiddleware(Routes(routes)).
		WithMiddleware(Headers()).
		WithMiddleware(Crypt()).
		WithMiddleware(Log()).
		Build()

	return conn
}

// NetworkConnection is the main entry point to the p2p network
type NetworkConnection struct {
	middlewares middlewares
	operators   []PeerOperator
	emitter     *eventEmitter
	peerStore   PeerStore
}

// Send will send the provided message to the given address
func (nc *NetworkConnection) Send(msg *Message, addr string) {
	nc.peerStore.LockPeer(addr, func(peer *Peer) {
		fmt.Printf("sending message: %s to peer %s\n", msg.PayloadGetString(), peer.RemoteAddress())
		peer.send <- msg
	})
}

// ConnectTo will Dial the provided peer by the given network
func (nc *NetworkConnection) ConnectTo(network string, addr string) {
	for _, op := range nc.operators {
		op.Dial(network, addr)
	}
}

// Start will start up the p2p network stack
func (nc *NetworkConnection) Start() error {
	nc.peerStore.OnPeerAdd(func(peer *Peer) {
		nc.emitter.EmitAsync("peer-new", peer)
	})
	nc.peerStore.OnPeerWantRemove(func(peer *Peer) {
		peer.stop()
		nc.peerStore.RemovePeer(peer)
	})

	for _, op := range nc.operators {
		op.OnPeer(func(a Adapter) {
			p := newPeer(a, nc.middlewares)
			err := nc.peerStore.AddPeer(p)
			if err != nil {
				p.emitter.EmitAsync("error", errors.Wrapf(err, "could not add peer: %s", p.RemoteAddress()))
				return
			}

			p.emitter.On("message", func(args []interface{}) {
				nc.emitter.EmitAsync("peer-message", args...)
			})
			p.emitter.On("disconnect", func(args []interface{}) {
				p := args[0].(*Peer)
				p.stop()
				nc.peerStore.RemovePeer(p)
				nc.emitter.EmitAsync("peer-disconnect", p)
			})
			p.emitter.On("error", func(args []interface{}) {
				p := args[0].(*Peer)
				err := args[1].(error)
				p.stop()
				nc.peerStore.RemovePeer(p)
				nc.emitter.EmitAsync("peer-error", p, err)
			})

			<-p.start()

			nc.emitter.EmitAsync("new-peer", p)
		})

		err := op.Start()
		if err != nil {
			return err
		}
	}

	nc.peerStore.Start()

	return nil
}

// OnPeer registers the provided handler and call it when a new peer connection is created
func (nc *NetworkConnection) OnPeer(handler func(p *Peer)) {
	nc.emitter.On("new-peer", func(args []interface{}) {
		handler(args[0].(*Peer))
	})
}

// OnMessage regsiters the given handler and call it when a new message is received
func (nc *NetworkConnection) OnMessage(handler func(p *Peer, msg *Message)) {
	nc.emitter.On("peer-message", func(args []interface{}) {
		handler(args[0].(*Peer), args[1].(*Message))
	})
}

// OnPeerError regsiters the given handler and call it when an error
// during the peer communication occurs
func (nc *NetworkConnection) OnPeerError(handler func(p *Peer, err error)) {
	nc.emitter.On("peer-error", func(args []interface{}) {
		handler(args[0].(*Peer), args[1].(error))
	})
}

// OnPeerDisconnect regsiters the given handler and call it when an the connection
// is lost
func (nc *NetworkConnection) OnPeerDisconnect(handler func(p *Peer)) {
	nc.emitter.On("peer-disconnect", func(args []interface{}) {
		handler(args[0].(*Peer))
	})
}

// Stop will shutdown the entire p2p network stack
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
