package go2p

import (
	"github.com/sirupsen/logrus"
)

/*
NewNetworkConnectionTCP provides a full configured TCP based network
It use the _DefaultMiddleware_ a TCP based operator and the following middleware:

Routes, Headers, Crypt, Log


*/
func NewNetworkConnectionTCP(localAddr string, routes RoutingTable) *NetworkConnection {
	op := NewTCPOperator("tcp", localAddr)

	conn := NewNetworkConnection().
		WithOperator(op).
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
	log         *logrus.Entry
	peers       *peers
}

// Send will send the provided message to the given address
func (nc *NetworkConnection) Send(msg *Message, addr string) {
	nc.peers.lock(addr, func(peer *Peer) {
		nc.log.WithFields(logrus.Fields{
			"local":  peer.LocalAddress(),
			"remote": peer.RemoteAddress(),
			"len":    len(msg.PayloadGet()),
		}).Debug("send messag")

		peer.send <- msg
	})
}

// SendBroadcast will send the given message to all peers
func (nc *NetworkConnection) SendBroadcast(msg *Message) {
	nc.peers.iteratePeer(func(peer *Peer) {
		nc.log.WithFields(logrus.Fields{
			"local":  peer.LocalAddress(),
			"remote": peer.RemoteAddress(),
			"len":    len(msg.PayloadGet()),
		}).Debug("send messag")

		peer.send <- msg
	})
}

// ConnectTo will Dial the provided peer by the given network
func (nc *NetworkConnection) ConnectTo(network string, addr string) {
	for _, op := range nc.operators {
		nc.log.WithFields(logrus.Fields{
			"network": network,
			"addr":    addr,
		}).Debug("dial peer")

		op.Dial(network, addr)
	}
}

// DisconnectFrom will disconnects the given peer
func (nc *NetworkConnection) DisconnectFrom(addr string) {
	nc.peers.lock(addr, func(peer *Peer) {
		nc.log.WithFields(logrus.Fields{
			"local":  peer.LocalAddress(),
			"remote": peer.RemoteAddress(),
		}).Debug("disconnect")

		peer.stop()
		go func(n *NetworkConnection, p *Peer) {
			n.peers.rm(p)
		}(nc, peer)
	})
}

// Start will start up the p2p network stack
func (nc *NetworkConnection) Start() error {
	nc.log.Debug("start network")

	// nc.peerStore.OnPeerAdd(func(peer *Peer) {
	// 	nc.emitter.EmitAsync("peer-new", peer)
	// })
	// nc.peerStore.OnPeerWantRemove(func(peer *Peer) {
	// 	peer.stop()
	// 	nc.peerStore.RemovePeer(peer)
	// })

	for _, op := range nc.operators {
		op.OnPeer(func(a Adapter) {
			p := newPeer(a, nc.middlewares)
			nc.peers.add(p)

			p.emitter.On("message", func(args []interface{}) {
				nc.emitter.EmitAsync("peer-message", args...)
			})
			p.emitter.On("disconnect", func(args []interface{}) {
				p := args[0].(*Peer)
				p.stop()
				nc.peers.rm(p)
				nc.emitter.EmitAsync("peer-disconnect", p)
			})
			p.emitter.On("error", func(args []interface{}) {
				p := args[0].(*Peer)
				err := args[1].(error)
				p.stop()
				nc.peers.rm(p)
				nc.emitter.EmitAsync("peer-error", p, err)
			})

			<-p.start()

			nc.emitter.EmitAsync("peer-connect", p)
		})

		err := op.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

// OnPeer registers the provided handler and call it when a new peer connection is created
func (nc *NetworkConnection) OnPeer(handler func(p *Peer)) {
	nc.emitter.On("peer-connect", func(args []interface{}) {
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

	nc.peers.iteratePeer(func(p *Peer) {
		nc.peers.rm(p)
		p.stop()
	})
}
