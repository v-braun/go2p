package go2p

import (
	"errors"
	"fmt"
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

	allActions       middlewares
	executingActions middlewares
	op               PipeOperation

	pos int // instruction pointer
}

func newPipe(peer *Peer, allActions middlewares, op PipeOperation, pos int, fromPos int, toPos int) *Pipe {
	if pos < fromPos {
		panic(fmt.Sprintf("invalid fromPos: %d is less than pos: %d", fromPos, pos))
	}
	if pos > toPos {
		panic(fmt.Sprintf("invalid toPos: %d is less than pos: %d", toPos, pos))
	}

	p := new(Pipe)

	p.op = op
	p.pos = pos
	p.allActions = allActions
	p.executingActions = allActions[fromPos:toPos]

	p.peer = peer

	return p
}
func (p *Pipe) process(msg *Message) error {
	nextItems := p.executingActions.nextItems(p.op)

	// fmt.Printf("next items for %v: %s \n", p.Operation(), nextItems.String())
	for _, m := range nextItems {
		fmt.Printf("exec msg: %s remote: %s local: %s action: %s direction: %v \n", msg.localID, p.peer.RemoteAddress(), p.peer.LocalAddress(), m.name, p.op)
		res, err := m.Execute(p.peer, p, msg)
		if err != nil {
			return err
		} else if res == Stop {
			return PipeStopProcessing
		}

		if p.op == Send {
			p.pos++
		} else {
			p.pos--
		}

	}

	return nil
}

func (p *Pipe) Send(msg *Message) error {
	pos := p.pos + 1
	from := p.pos + 1
	to := len(p.allActions)

	if pos > to {
		err := p.peer.io.sendMsg(msg)
		return err
	}

	subPipe := newPipe(p.peer, p.allActions, Send, pos, from, to)

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
		pos := p.pos + 1
		from := p.pos + 1
		to := len(p.allActions)
		if pos <= to {
			subPipe := newPipe(p.peer, p.allActions, Receive, pos, from, to)
			err = subPipe.process(msg)
		}
	}

	return msg, err
}

func (p *Pipe) Operation() PipeOperation {
	return p.op
}
