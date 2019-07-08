package go2p

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
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

// ErrPipeStopProcessing is returned when the pipe has stopped it execution
var ErrPipeStopProcessing = errors.New("pipe stopped")

// Pipe handles the processing of an message
type Pipe struct {
	peer *Peer

	allActions       middlewares
	executingActions middlewares
	op               PipeOperation

	pos int // instruction pointer

	log *logrus.Entry
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
	p.log = newLogger("pipe")

	p.peer = peer

	return p
}
func (p *Pipe) process(msg *Message) error {
	nextItems := p.executingActions.nextItems(p.op)

	for _, m := range nextItems {
		p.log.WithFields(logrus.Fields{
			"name":    m.name,
			"pos":     m.pos,
			"msg-len": len(msg.PayloadGet()),
		}).Debug("execute middleware")

		res, err := m.execute(p.peer, p, msg)
		if err != nil {
			p.log.WithFields(logrus.Fields{
				"name":    m.name,
				"pos":     m.pos,
				"msg-len": len(msg.PayloadGet()),
				"err":     err,
			}).Error("middleware error")
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

// Receive will block the current call until a message was read from the peer or
// an error occurs.
//
// The message goes only through middlewares that are after the current pipe position
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

// Operation returns the current pipe operation (Send or Receive)
func (p *Pipe) Operation() PipeOperation {
	return p.op
}
