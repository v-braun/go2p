package tcp

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/core/utils"
)

var _ core.Operator = (*operator)(nil)

type operator struct {
	emitter *utils.EventEmitter
	server  net.Listener
	ctx     context.Context
	cancel  context.CancelFunc

	localNetwok string
	localAddr   string
}

// NewOperator creates a new TCP based PeerOperator instance
func NewOperator(network string, localAddr string) core.Operator {
	o := new(operator)
	o.emitter = utils.NewEventEmitter()
	o.localNetwok = network
	o.localAddr = localAddr
	return o
}

// Dial connects to the address by the given network
func (o *operator) Dial(network string, addr string) error {
	if network != "tcp" {
		return core.ErrInvalidNetwork
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return err
	}

	c := NewConn(conn)
	o.emitter.Emit("new-peer", c)
	return nil
}

// OnPeer registers the given handler and calls it when a new peer connection is
// established
func (o *operator) OnPeer(handler func(p core.Conn)) {
	o.emitter.On("new-peer", func(p core.Conn) {
		handler(p)
	})
}

// OnError registers the given handler and calls it when a peer error occurs
func (o *operator) OnError(handler func(err error)) {
	o.emitter.On("error", func(err error) {
		handler(err)
	})
}

// Start will start the net.Listener and waits for incoming connections
func (o *operator) Start() error {
	if o.localNetwok != "tcp" {
		return core.ErrInvalidNetwork
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
func (o *operator) Stop() {
	o.cancel()
	o.server.Close()
}

func (o *operator) listen(ctx context.Context) {
	go (func(o *operator, ctx context.Context) {
		for {
			conn, err := o.server.Accept()
			if err == nil && conn != nil {
				c := NewConn(conn)
				o.emitter.Emit("new-peer", c)
			} else if tmpErr, ok := err.(net.Error); ok && tmpErr.Temporary() {
				o.emitter.Emit("error", errors.Wrap(err, "temp error during listening"), true)
			} else if err != nil && ctx.Err() == nil {
				o.emitter.Emit("error", errors.Wrap(err, "fatal error, will stop listening"))
				break
			} else if ctx.Err() != nil {
				break
			}
		}

	})(o, ctx)
}
