package go2p

import (
	"errors"
	"fmt"

	"github.com/v-braun/go2p/rsa_utils"
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

	fmt.Printf("next items for %v: %s \n", p.Operation(), nextItems.String())
	for _, m := range nextItems {
		fmt.Printf("%s | %s [%v] %s (%d) pipePos: %d dara: {%s} (%d) \n", msg.localID, p.peer.Address(), p.Operation(), m.name, m.pos, p.pos, rsa_utils.PrintableStr(msg.PayloadGet(), 10), len(msg.PayloadGet()))
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

	fmt.Printf("send %s... to %s\n", rsa_utils.PrintableStr(msg.PayloadGet(), 8), p.peer.Address())
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
		fmt.Printf("received %s... from %s\n", rsa_utils.PrintableStr(msg.PayloadGet(), 8), p.peer.Address())

		subPipe := newPipe(p.peer, p.allActions, Receive, p.pos+1)
		err = subPipe.process(msg)
	}

	return msg, err
}

func (p *Pipe) Operation() PipeOperation {
	return p.op
}
