package core

import (
	"io"

	"github.com/pkg/errors"
	"github.com/v-braun/awaiter"
	"github.com/v-braun/go2p/core/utils"
)

type connHost struct {
	Conn
	receive chan *Message
	send    chan *Message

	awaiter awaiter.Awaiter

	emitter *utils.EventEmitter
}

func newConnHost(conn Conn) *connHost {
	host := new(connHost)
	host.receive = make(chan *Message)
	host.send = make(chan *Message)
	host.awaiter = awaiter.New()
	host.Conn = conn
	host.emitter = utils.NewEventEmitter()

	return host
}

func (host *connHost) start() {
	host.awaiter.Go(func() {
		for {
			m, err := host.ReadMessage()
			if err != nil {
				host.handleError(err, "read")
				return
			}

			select {
			case host.receive <- m:
				continue
			case <-host.awaiter.CancelRequested():
				return
			}
		}
	})

	host.awaiter.Go(func() {
		for {
			select {
			case m := <-host.send:
				err := host.WriteMessage(m)
				if err != nil {
					host.handleError(err, "write")
					return
				}

				continue
			case <-host.awaiter.CancelRequested():
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
func (host *connHost) handleError(err error, src string) {
	if isDisconnectErr(err) {
		host.emitter.Emit("disconnect")
		return
	}

	host.emitter.Emit("error", errors.Wrapf(err, "error during %s", src))
}

func (host *connHost) sendMsg(m *Message) error {
	select {
	case host.send <- m:
		return nil
	case <-host.awaiter.CancelRequested():
		return DisconnectedError
	}
}

func (host *connHost) receiveMsg() (*Message, error) {
	select {
	case m := <-host.receive:
		return m, nil
	case <-host.awaiter.CancelRequested():
		return nil, DisconnectedError
	}
}
