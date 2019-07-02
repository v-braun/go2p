package go2p

import (
	"errors"
)

type PipeOperation int

const (
	Send    PipeOperation = iota
	Receive PipeOperation = iota
)

func (po PipeOperation) String() string {
	return [...]string{"Send", "Receive"}[po]
}

var PipeStopProcessing = errors.New("pipe stopped")

type Pipe struct {
	peer *Peer

	allActions middlewares
	op         PipeOperation

	pos int // instruction pointer
}

func newPipe(peer *Peer, allActions middlewares, op PipeOperation, pos int) *Pipe {
	p := new(Pipe)

	p.op = op
	p.pos = pos
	p.allActions = allActions

	p.peer = peer

	return p
}
func (p *Pipe) process(msg *Message) error {
	nextItems := p.allActions.nextItems(p.op, p.pos)

	for _, m := range nextItems {
		// fmt.Printf("%s | %s [%v] %s \n", msg.localId, p.peer.Address(), p.Operation(), m.name)
		res, err := m.Execute(p.peer, p, msg)
		if err != nil {
			return err
		} else if res == Stop {
			return PipeStopProcessing
		}

		p.pos++
	}

	return nil
}

func (p *Pipe) Send(msg *Message) error {
	subPipe := newPipe(p.peer, p.allActions, Send, p.pos+1)

	if err := subPipe.process(msg); err != nil {
		return err
	}

	err := p.peer.io.sendMsg(msg)
	return err
}

func (p *Pipe) Receive() (*Message, error) {
	msg, err := p.peer.io.receiveMsg()
	if err == nil && msg == nil {
		panic("unexpected nil result from peer.receive")
	} else if err != nil {
		return nil, err
	} else {
		subPipe := newPipe(p.peer, p.allActions, Receive, p.pos+1)
		err = subPipe.process(msg)
	}

	return msg, err
}

func (p *Pipe) Operation() PipeOperation {
	return p.op
}
