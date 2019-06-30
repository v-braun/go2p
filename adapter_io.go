package go2p

import (
	"context"
	"io"
	"sync"

	"github.com/olebedev/emitter"
	"github.com/pkg/errors"
)

type adapterIO struct {
	receive chan *Message
	send    chan *Message
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	adapter Adapter

	ctx context.Context

	emitter *emitter.Emitter
}

func newAdapterIO(adapter Adapter) *adapterIO {
	io := new(adapterIO)
	io.receive = make(chan *Message)
	io.send = make(chan *Message)
	io.ctx, io.cancel = context.WithCancel(context.Background())
	io.wg = sync.WaitGroup{}
	io.adapter = adapter
	io.emitter = emitter.New(10)
	io.emitter.Use("*", emitter.Void)

	return io
}

func (io *adapterIO) start() {
	io.wg.Add(1)
	go func(io *adapterIO) {
		defer io.wg.Done()

		for {
			m, err := io.adapter.ReadMessage()
			if err != nil {
				io.handleError(err, "read")
				return
			}

			select {
			case io.receive <- m:
				continue
			case <-io.ctx.Done():
				return
			}
		}
	}(io)

	io.wg.Add(1)
	go func(io *adapterIO) {
		defer io.wg.Done()

		for {
			select {
			case m := <-io.send:
				err := io.adapter.WriteMessage(m)
				if err != nil {
					io.handleError(err, "write")
					return
				}

				continue
			case <-io.ctx.Done():
				return
			}
		}
	}(io)
}

func isDisconnectErr(err error) bool {
	return err == DisconnectedError || err == io.EOF
}
func (io *adapterIO) handleError(err error, src string) {
	if isDisconnectErr(err) {
		go io.emitter.Emit("disconnect")
		return
	}

	io.emitter.Emit("error", errors.Wrapf(err, "error during %s", src))
}

func (io *adapterIO) sendMsg(m *Message) error {
	select {
	case io.send <- m:
		return nil
	case <-io.ctx.Done():
		return DisconnectedError
	}
}

func (io *adapterIO) receiveMsg() (*Message, error) {
	select {
	case m := <-io.receive:
		return m, nil
	case <-io.ctx.Done():
		return nil, DisconnectedError
	}
}
