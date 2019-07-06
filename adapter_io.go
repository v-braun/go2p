package go2p

import (
	"io"

	"github.com/pkg/errors"
	"github.com/v-braun/awaiter"
)

type adapterIO struct {
	receive chan *Message
	send    chan *Message

	awaiter awaiter.Awaiter

	adapter Adapter

	emitter *eventEmitter
}

func newAdapterIO(adapter Adapter) *adapterIO {
	io := new(adapterIO)
	io.receive = make(chan *Message)
	io.send = make(chan *Message)
	io.awaiter = awaiter.New()
	io.adapter = adapter
	io.emitter = newEventEmitter()

	return io
}

func (io *adapterIO) start() {
	io.awaiter.Go(func() {
		for {
			m, err := io.adapter.ReadMessage()
			if err != nil {
				io.handleError(err, "read")
				return
			}

			select {
			case io.receive <- m:
				continue
			case <-io.awaiter.CancelRequested():
				return
			}
		}
	})

	io.awaiter.Go(func() {
		for {
			select {
			case m := <-io.send:
				err := io.adapter.WriteMessage(m)
				if err != nil {
					io.handleError(err, "write")
					return
				}

				continue
			case <-io.awaiter.CancelRequested():
				return
			}
		}
	})
}

func isDisconnectErr(err error) bool {
	if err == DisconnectedError || err == io.EOF {
		return true
	}

	return false
}
func (io *adapterIO) handleError(err error, src string) {
	if isDisconnectErr(err) {
		io.emitter.EmitAsync("disconnect")
		return
	}

	io.emitter.EmitAsync("error", errors.Wrapf(err, "error during %s", src))
}

func (io *adapterIO) sendMsg(m *Message) error {
	select {
	case io.send <- m:
		return nil
	case <-io.awaiter.CancelRequested():
		return DisconnectedError
	}
}

func (io *adapterIO) receiveMsg() (*Message, error) {
	select {
	case m := <-io.receive:
		return m, nil
	case <-io.awaiter.CancelRequested():
		return nil, DisconnectedError
	}
}
