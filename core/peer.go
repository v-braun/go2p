package core

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/v-braun/go2p/must"

	"github.com/pkg/errors"
)

type Peer interface {
	LocalAddr() string
	RemoteAddr() string
	Connected() bool
}

var _ Peer = (*peer)(nil)

type peer struct {
	adapter Adapter

	receiveMsg chan Message
	receiveErr chan error

	stop      chan struct{}
	connected bool

	lock sync.Mutex
	wg   sync.WaitGroup
}

func (p *peer) LocalAddr() string {
	return p.LocalAddr()
}

func (p *peer) RemoteAddr() string {
	return p.RemoteAddr()
}

func (p *peer) Connected() bool {
	select {
	case <-p.stop:
		return false
	default:
		return true
	}
}

func newPeer(adapter Adapter) *peer {
	must.ArgNotNil(adapter, "adapter")

	result := new(peer)
	result.adapter = adapter

	result.receiveMsg = make(chan Message)
	result.receiveErr = make(chan error)

	result.connected = true
	result.stop = make(chan struct{})

	result.lock = sync.Mutex{}
	result.wg = sync.WaitGroup{}

	return result
}

func (p *peer) run() {
	p.wg.Add(1)
	go p.receiveLoop()
}

func (p *peer) receiveLoop() {
	defer func() {
		close(p.receiveMsg)
		close(p.receiveErr)
		defer p.wg.Done()
	}()

	for {
		msg, err := p.readMessage()
		if err == DisconnectedError {
			return
		} else if err != nil {
			select {
			case <-p.stop:
				return
			case p.receiveErr <- err:
				return
			}
		} else if msg != nil {
			select {
			case <-p.stop:
				return
			case p.receiveMsg <- msg:
			}
		}
	}
}

func (p *peer) send(msg Message) error {
	err := p.writeMessage(msg)

	return err
}

func (p *peer) receive() (Message, error) {
	select {
	case err := <-p.receiveErr:
		return nil, err
	case msg := <-p.receiveMsg:
		return msg, nil
	case <-p.stop:
		return nil, DisconnectedError

	}
}

func (p *peer) shutdown() {

	p.lock.Lock()
	defer p.lock.Unlock()

	if p.connected {
		p.connected = false
		close(p.stop)
		p.adapter.Close()
		p.wg.Wait()
	}
}

func (p *peer) readMessage() (Message, error) {
	sizeBuffer := make([]byte, 4)
	if err := p.readBuffer(len(sizeBuffer), sizeBuffer, "size"); err != nil {
		return nil, err
	}

	size := int(binary.BigEndian.Uint32(sizeBuffer))
	payloadBuffer := make([]byte, size)
	if err := p.readBuffer(len(payloadBuffer), payloadBuffer, "payload"); err != nil {
		return nil, err
	}

	msg := new(message)
	msg.payload = payloadBuffer

	return msg, nil
}

func (p *peer) writeMessage(msg Message) error {
	must.ArgNotNil(msg, "msg")

	payload := msg.PayloadGet()

	size := uint32(len(payload))
	sizeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuffer, size)

	fullPayload := append(sizeBuffer, payload...)

	err := p.writeBuffer(fullPayload)
	return err
}

func isConnEnd(err error) bool {
	return err == DisconnectedError || err == io.EOF
}

func (p *peer) readBuffer(length int, buffer []byte, dataType string) error {
	var readed int
	for readed < length {
		currentReaded, err := p.adapter.Read(buffer[readed:])
		if isConnEnd(err) {
			return DisconnectedError
		} else if err != nil {
			return errors.Wrapf(err, "failed read message %s", dataType)
		} else {
			readed += currentReaded
		}
	}

	return nil
}
func (p *peer) writeBuffer(buffer []byte) error {

	err := p.adapter.Write(buffer)
	if isConnEnd(err) {
		return DisconnectedError
	} else if err != nil {
		return errors.Wrapf(err, "failed write message")
	} else {
		return nil
	}
}
