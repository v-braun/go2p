package go2p

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
)

func TestMiddlewareError(t *testing.T) {
	f := func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
		return Stop, errors.New("fail")
	}

	p := newPipe(nil, newMiddlewares(NewMiddleware("fail", f)), Send, 0, 0, 1)
	err := p.process(NewMessage())
	assert.Error(t, err)
}
