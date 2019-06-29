package go2p

import (
	"fmt"
)

type PipeOperation int

const (
	Send    PipeOperation = iota
	Receive PipeOperation = iota
)

type pipe struct {
	peer *Peer

	allActions middlewares
	op         PipeOperation

	pos int // instruction pointer
}

func newPipe(peer *Peer, allActions middlewares, op PipeOperation, pos int) *pipe {
	p := new(pipe)

	p.op = op
	p.pos = pos
	p.allActions = allActions

	p.peer = peer

	return p
}
func (p *pipe) process(msg *Message) error {
	nextItems := p.allActions.nextItems(p.op, p.pos)

	for _, m := range nextItems {
		fmt.Println(m.name)
		res, err := m.Execute(p.peer, msg)
		if err != nil {
			p.handleErr(err)
			return err
		} else if res == Stop {
			return nil
		}

		p.pos++
	}

	return nil
}

func (p *pipe) handleErr(err error) error {
	if err == nil {
		return nil
	}

	panic("todo: implement")
	// p.peer.shutdown()
	// notify

	return err
}

func (p *pipe) Send(msg *Message) error {
	subPipe := newPipe(p.peer, p.allActions, Send, p.pos+1)

	if err := subPipe.process(msg); err != nil {
		return p.handleErr(err)
	}

	err := p.peer.adapter.Send(msg)
	err = p.handleErr(err)
	return err
}

func (p *pipe) Receive() (*Message, error) {
	msg, err := p.peer.adapter.Receive()
	if err == nil && msg == nil {
		panic("unexpected nil result from peer.receive")
	} else if err != nil {
		return nil, p.handleErr(err)
	} else {
		subPipe := newPipe(p.peer, p.allActions, Receive, p.pos+1)
		err = subPipe.process(msg)
	}

	err = p.handleErr(err)

	return msg, err
}

func (p *pipe) Operation() PipeOperation {
	return p.op
}
