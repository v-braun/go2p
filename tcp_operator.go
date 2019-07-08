package go2p

import (
	"context"
	"net"

	"github.com/pkg/errors"
)

var _ PeerOperator = (*OperatorTCP)(nil)

// OperatorTCP is an implementation of the PeerOperator inteface that handles
// TCP based connections.
// It use an net.Listener for incoming connections and and tcp.Dialer for outgoing.
type OperatorTCP struct {
	emitter *eventEmitter
	server  net.Listener
	ctx     context.Context
	cancel  context.CancelFunc

	localNetwok string
	localAddr   string
}

// NewTCPOperator creates a new TCP based PeerOperator instance
func NewTCPOperator(network string, localAddr string) *OperatorTCP {
	o := new(OperatorTCP)
	o.emitter = newEventEmitter()
	o.localNetwok = network
	o.localAddr = localAddr
	return o
}

// Dial connectes to the address by the given network
func (o *OperatorTCP) Dial(network string, addr string) error {
	if network != "tcp" {
		return ErrInvalidNetwork
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return err
	}

	adapter := NewAdapter(conn)
	o.emitter.EmitAsync("new-peer", adapter)
	return nil
}

// OnPeer registers the given handler and calls it when a new peer connection is
// established
func (o *OperatorTCP) OnPeer(handler func(p Adapter)) {
	o.emitter.On("new-peer", func(args []interface{}) {
		handler(args[0].(Adapter))
	})
}

// OnError registers the given handler and calls it when a peer error occurs
func (o *OperatorTCP) OnError(handler func(err error)) {
	o.emitter.On("error", func(args []interface{}) {
		handler(args[0].(error))
	})
}

// Start will start the net.Listener and waits for incoming connections
func (o *OperatorTCP) Start() error {
	if o.localNetwok != "tcp" {
		return ErrInvalidNetwork
	}

	listener, err := net.Listen(o.localNetwok, o.localAddr)
	if err != nil {
		return err
	}

	o.ctx, o.cancel = context.WithCancel(context.Background())

	o.server = listener
	go o.listen(o.ctx)
	return nil
}

// Stop will close the underlining net.Listener
func (o *OperatorTCP) Stop() {
	o.cancel()
	o.server.Close()
}

func (o *OperatorTCP) listen(ctx context.Context) {
	go (func(o *OperatorTCP, ctx context.Context) {
		for {
			conn, err := o.server.Accept()
			if err == nil && conn != nil {
				adapter := NewAdapter(conn)
				o.emitter.EmitAsync("new-peer", adapter)
			} else if tmpErr, ok := err.(net.Error); ok && tmpErr.Temporary() {
				o.emitter.EmitAsync("error", errors.Wrap(err, "temp error during listening"), true)
				continue
			} else if err != nil && ctx.Err() == nil {
				o.emitter.EmitAsync("error", errors.Wrap(err, "fatal error, wil stop listening"))
				break
			} else if ctx.Err() != nil {
				break
			}
		}

	})(o, ctx)
}
