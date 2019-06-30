package go2p

import (
	"context"
	"net"

	"github.com/olebedev/emitter"
	"github.com/pkg/errors"
)

var _ PeerOperator = (*OperatorTCP)(nil)

type OperatorTCP struct {
	emitter *emitter.Emitter
	server  net.Listener
	ctx     context.Context
	cancel  context.CancelFunc

	localNetwok string
	localAddr   string
}

func NewTcpOperator(network string, localAddr string) *OperatorTCP {
	o := new(OperatorTCP)
	o.emitter = emitter.New(10)
	o.emitter.Use("*", emitter.Void)
	o.localNetwok = network
	o.localAddr = localAddr
	return o
}

func (o *OperatorTCP) Dial(network string, addr string) error {
	if network != "tcp" {
		return InvalidNetworkError
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return err
	}

	adapter := NewAdapter(conn)
	o.emitter.Emit("new-peer", adapter)
	return nil
}

func (o *OperatorTCP) OnPeer(handler func(p Adapter)) {
	o.emitter.On("new-peer", func(ev *emitter.Event) {
		handler(ev.Args[0].(Adapter))
	})
}

func (o *OperatorTCP) OnError(handler func(err error)) {
	o.emitter.On("error", func(ev *emitter.Event) {
		handler(ev.Args[0].(error))
	})
}

func (o *OperatorTCP) Start() error {
	if o.localNetwok != "tcp" {
		return InvalidNetworkError
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
				o.emitter.Emit("new-peer", adapter)
			} else if tmpErr, ok := err.(net.Error); ok && tmpErr.Temporary() {
				o.emitter.Emit("error", errors.Wrap(err, "temp error during listening"), true)
				continue
			} else if err != nil && ctx.Err() == nil {
				o.emitter.Emit("error", errors.Wrap(err, "fatal error, wil stop listening"))
				break
			} else if ctx.Err() != nil {
				break
			}
		}

	})(o, ctx)
}
