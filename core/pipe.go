package core

import (
	"fmt"
)

type Pipe interface {
	Send(msg Message) error
	Receive() (Message, error)
	Peer() Peer
	Operation() PipeOperation
}

var _ Pipe = (*pipe)(nil)

type pipe struct {
	peer *peer

	store PeerStore

	allActions middlewares
	op         PipeOperation

	ip int // instruction pointer
}

func newPipe(peer *peer, allActions middlewares, op PipeOperation, pos int) *pipe {
	p := new(pipe)

	p.op = op
	p.ip = pos
	p.allActions = allActions

	p.peer = peer

	return p
}
func (p *pipe) process(msg Message) error {
	nextItems := p.allActions.nextItems(p.op, p.ip)

	for _, m := range nextItems {
		fmt.Println(m.name)
		res, err := m.Execute(p, msg)
		if err != nil {
			p.handleErr(err)
			return err
		} else if res == Stop {
			return nil
		}

		p.ip++
	}

	return nil
}

func (p *pipe) handleErr(err error) error {
	if err == nil {
		return nil
	}

	p.peer.shutdown()
	// notify

	return err
}

func (p *pipe) Send(msg Message) error {
	subPipe := newPipe(p.peer, p.allActions, Send, p.ip+1)

	if err := subPipe.process(msg); err != nil {
		return p.handleErr(err)
	}

	err := p.peer.send(msg)
	if err != nil {
		return p.handleErr(err)
	}

	return nil
}

func (p *pipe) Receive() (Message, error) {
	msg, err := p.peer.receive()
	if err == nil && msg == nil {
		panic("unexpected nil result from peer.receive")
	} else if err != nil {
		return nil, p.handleErr(err)
	} else {
		subPipe := newPipe(p.peer, p.allActions, Receive, p.ip+1)
		err = subPipe.process(msg)
	}

	return msg, p.handleErr(err)
}

func (p *pipe) Peer() Peer {
	return p.peer
}

func (p *pipe) Operation() PipeOperation {
	return p.op
}
