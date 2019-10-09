package core

import (
	"github.com/v-braun/go2p/core/logging"
)

// PipeOperation represents the pipe direction (Send or Receive)
type PipeOperation int

const (
	// Send represents an outgoing message pipe processing
	Send PipeOperation = iota
	// Receive represents an incoming message pipe processing
	Receive PipeOperation = iota
)

func (po PipeOperation) String() string {
	return [...]string{"Send", "Receive"}[po]
}

// Pipe handles the processing of an message
type Pipe struct {
	peer *Peer

	allActions       middlewares
	executingActions middlewares
	op               PipeOperation

	pos int // instruction pointer

	log *logging.Logger
}

func newPipe(peer *Peer, allActions middlewares, op PipeOperation, pos int, fromPos int, toPos int) *Pipe {
	p := new(Pipe)

	p.op = op
	p.pos = pos
	p.allActions = allActions
	p.executingActions = allActions[fromPos:toPos]
	p.log = logging.NewLogger("pipe")

	p.peer = peer

	return p
}
func (p *Pipe) process(msg *Message) error {
	nextItems := p.executingActions.nextItems(p.op)

	for _, m := range nextItems {
		p.log.Debug(logging.Fields{
			"name":    m.name,
			"pos":     m.pos,
			"msg-len": len(msg.PayloadGet()),
			"op":      p.op.String(),
		}, "execute middleware")

		res, err := m.execute(p.peer, p, msg)
		if err != nil {
			p.log.Error(logging.Fields{
				"name":    m.name,
				"pos":     m.pos,
				"msg-len": len(msg.PayloadGet()),
				"err":     err,
			}, "middleware error")
			return err
		} else if res == Stop {
			return ErrPipeStopProcessing
		}

		if p.op == Send {
			p.pos++
		} else {
			p.pos--
		}

	}

	return nil
}

// Send will send the provided message during the current pipe execution.
//
// The message goes only through middlewares that are after the current pipe position
func (p *Pipe) Send(msg *Message) error {
	pos := p.pos + 1
	from := p.pos + 1
	to := len(p.allActions)

	if pos > to {
		err := p.peer.conn.sendMsg(msg)
		return err
	}

	subPipe := newPipe(p.peer, p.allActions, Send, pos, from, to)

	if err := subPipe.process(msg); err != nil {
		return err
	}

	err := p.peer.conn.sendMsg(msg)
	return err
}

// Receive will block the current call until a message was read from the peer or
// an error occurs.
//
// The message goes only through middlewares that are after the current pipe position
func (p *Pipe) Receive() (*Message, error) {
	msg, err := p.peer.conn.receiveMsg()
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

// Operation returns the current pipe operation (Send or Receive)
func (p *Pipe) Operation() PipeOperation {
	return p.op
}
