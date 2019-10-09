package core

import (
	"github.com/v-braun/go2p/core/logging"
	"github.com/v-braun/go2p/core/utils"
)

// Network is the main entry point to the p2p network
type Network struct {
	middlewares middlewares
	operators   []Operator
	emitter     *utils.EventEmitter
	log         *logging.Logger
	peers       *peers
	started     bool
}

// NewNetwork will create a new empty network
func NewNetwork() *Network {
	result := &Network{
		middlewares: make([]*Middleware, 0),
		operators:   make([]Operator, 0),
		emitter:     utils.NewEventEmitter(),
		log:         logging.NewLogger("network"),
		peers:       newPeers(),
		started:     false,
	}

	return result
}

func (nc *Network) UseMiddleware(middleware *Middleware) *Network {
	nc.ensureStarted(false)
	middlewares := append(nc.middlewares, middleware)
	nc.middlewares = newMiddlewares(middlewares...)
	return nc
}

func (nc *Network) UseOperator(operator Operator) *Network {
	nc.ensureStarted(false)

	nc.operators = append(nc.operators, operator)
	return nc
}

// Send will send the provided message to the given address
func (nc *Network) Send(msg *Message, addr string) {
	nc.ensureStarted(true)

	peer := nc.peers.findByAddr(addr)
	if peer != nil {
		nc.log.Debug(logging.Fields{
			"local":  peer.LocalAddress(),
			"remote": peer.RemoteAddress(),
			"len":    len(msg.PayloadGet()),
		}, "send messag")
		peer.send <- msg
	}
}

// SendBroadcast will send the given message to all peers
func (nc *Network) SendBroadcast(msg *Message) {
	nc.ensureStarted(true)

	peers := nc.peers.allPeers()
	for _, peer := range peers {
		nc.log.Debug(logging.Fields{
			"local":  peer.LocalAddress(),
			"remote": peer.RemoteAddress(),
			"len":    len(msg.PayloadGet()),
		}, "send messag")

		copy := msg.Clone()
		peer.send <- copy
	}
}

// ConnectTo will Dial the provided peer by the given network
func (nc *Network) ConnectTo(network string, addr string) error {
	nc.ensureStarted(true)

	for _, op := range nc.operators {
		nc.log.Debug(logging.Fields{
			"network": network,
			"addr":    addr,
		}, "dial peer")

		err := op.Dial(network, addr)
		if err != nil && err != ErrInvalidNetwork {
			return err
		}
	}

	return nil
}

// DisconnectFrom will disconnects the given peer
func (nc *Network) DisconnectFrom(addr string) {
	nc.ensureStarted(true)

	peer := nc.peers.findByAddr(addr)
	if peer != nil {
		nc.log.Debug(logging.Fields{
			"local":  peer.LocalAddress(),
			"remote": peer.RemoteAddress(),
		}, "disconnect")

		peer.stop()
		go func(n *Network, p *Peer) {
			n.peers.rm(p)
		}(nc, peer)
	}
}

// Start will start up the p2p network stack
func (nc *Network) Start() error {
	nc.ensureStarted(false)

	nc.log.Debug(logging.Fields{}, "start network")

	for _, op := range nc.operators {
		op.OnPeer(func(a Conn) {
			p := newPeer(a, nc.middlewares)
			nc.peers.add(p)

			p.emitter.On("message", func(p *Peer, m *Message) {
				nc.emitter.Emit("peer-message", p, m)
			})
			p.emitter.On("disconnect", func(p *Peer) {
				p.stop()
				nc.peers.rm(p)
				nc.emitter.Emit("peer-disconnect", p)
			})
			p.emitter.On("error", func(p *Peer, err error) {
				p.stop()
				nc.peers.rm(p)
				nc.emitter.Emit("peer-error", p, err)
			})

			<-p.start()

			nc.emitter.Emit("peer-connect", p)
		})

		err := op.Start()
		if err != nil {
			return err
		}
	}

	nc.started = true

	return nil
}

// OnPeer registers the provided handler and call it when a new peer connection is created
func (nc *Network) OnPeer(handler func(p *Peer)) {
	nc.ensureStarted(false)

	nc.emitter.On("peer-connect", func(p *Peer) {
		handler(p)
	})
}

// OnMessage regsiters the given handler and call it when a new message is received
func (nc *Network) OnMessage(handler func(p *Peer, msg *Message)) {
	nc.ensureStarted(false)

	nc.emitter.On("peer-message", func(p *Peer, msg *Message) {
		handler(p, msg)
	})
}

// OnPeerError regsiters the given handler and call it when an error
// during the peer communication occurs
func (nc *Network) OnPeerError(handler func(p *Peer, err error)) {
	nc.ensureStarted(false)

	nc.emitter.On("peer-error", func(p *Peer, err error) {
		handler(p, err)
	})
}

// OnPeerDisconnect regsiters the given handler and call it when an the connection
// is lost
func (nc *Network) OnPeerDisconnect(handler func(p *Peer)) {
	nc.ensureStarted(false)

	nc.emitter.On("peer-disconnect", func(p *Peer) {
		handler(p)
	})
}

// Stop will shutdown the entire p2p network stack
func (nc *Network) Stop() {
	nc.ensureStarted(true)

	for _, op := range nc.operators {
		op.Stop()
	}

	peers := nc.peers.allPeers()
	for _, peer := range peers {
		nc.peers.rm(peer)
		peer.stop()
	}

	nc.started = false
}

func (nc *Network) ensureStarted(state bool) {
	if nc.started == state {
		return
	}

	if state {
		panic("invalid operation: network must be started")
	} else {
		panic("invalid operation: network is already started")
	}
}
