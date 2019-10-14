package tcp

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/core/utils"
)

var _ core.Extension = (*extension)(nil)
var _ core.Dialer = (*extension)(nil)
var _ core.Listener = (*extension)(nil)

type extension struct {
	emitter *utils.EventEmitter
	server  net.Listener
	ctx     context.Context
	cancel  context.CancelFunc

	localNetwok string
	localAddr   string
}

// NewOperator creates a new TCP based PeerOperator instance
func NewTCPExtension(network string, localAddr string) core.Extension {
	ex := new(extension)
	ex.emitter = utils.NewEventEmitter()
	ex.localNetwok = network
	ex.localAddr = localAddr
	return ex
}

// Dial connects to the address by the given network
func (ex *extension) Dial(network string, addr string) error {
	if network != "tcp" {
		return core.ErrInvalidNetwork
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return err
	}

	c := NewConn(conn)
	ex.emitter.Emit("new-peer", c)
	return nil
}

// OnPeer registers the given handler and calls it when a new peer connection is
// established
func (ex *extension) OnPeer(handler func(p core.Conn)) {
	ex.emitter.On("new-peer", func(p core.Conn) {
		handler(p)
	})
}

// OnError registers the given handler and calls it when a peer error occurs
func (ex *extension) OnError(handler func(err error)) {
	ex.emitter.On("error", func(err error) {
		handler(err)
	})
}

// Start will start the net.Listener and waits for incoming connections
func (ex *extension) Install(nc *core.Network) error {
	if ex.localNetwok != "tcp" {
		return core.ErrInvalidNetwork
	}

	listener, err := net.Listen(ex.localNetwok, ex.localAddr)
	if err != nil {
		return err
	}

	ex.ctx, ex.cancel = context.WithCancel(context.Background())

	ex.server = listener
	go ex.listen(ex.ctx)
	return nil
}

// Stop will close the underlining net.Listener
func (ex *extension) Uninstall() {
	ex.cancel()
	ex.server.Close()
}

func (ex *extension) listen(ctx context.Context) {
	go (func(ex *extension, ctx context.Context) {
		for {
			conn, err := ex.server.Accept()
			if err == nil && conn != nil {
				c := NewConn(conn)
				ex.emitter.Emit("new-peer", c)
			} else if tmpErr, ok := err.(net.Error); ok && tmpErr.Temporary() {
				ex.emitter.Emit("error", errors.Wrap(err, "temp error during listening"), true)
			} else if err != nil && ctx.Err() == nil {
				ex.emitter.Emit("error", errors.Wrap(err, "fatal error, will stop listening"))
				break
			} else if ctx.Err() != nil {
				break
			}
		}

	})(ex, ctx)
}
