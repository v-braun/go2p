package go2p

import (
	"fmt"
	"sort"
)

type MiddlewareResult int

const (
	Stop MiddlewareResult = iota
	Next MiddlewareResult = iota
)

type MiddlewareFunc func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error)

type Middleware struct {
	Execute MiddlewareFunc
	name    string
	pos     int
}

func NewMiddleware(name string, action MiddlewareFunc) *Middleware {
	return &Middleware{
		name:    name,
		Execute: action,
	}
}

func (self *Middleware) Pos() int {
	return self.pos
}

func (self *Middleware) String() string {
	return fmt.Sprintf("%s (%d)", self.name, self.pos)
}

type middlewares []*Middleware

func newMiddlewares(actions ...*Middleware) middlewares {
	result := middlewares{}
	for idx, action := range actions {
		action.pos = idx
		result = append(result, action)
	}

	return result
}

func (ml middlewares) nextItems(op PipeOperation) middlewares {
	result := ml.Copy()
	sort.Sort(result)
	if op == Receive {
		sort.Sort(sort.Reverse(result))
	}

	return result
}

func (ml middlewares) Copy() middlewares {
	result := make(middlewares, len(ml))
	copy(result, ml)
	return result
}

func (ml middlewares) Len() int {
	return len(ml)
}

func (ml middlewares) Swap(i, j int) {
	ml[i], ml[j] = ml[j], ml[i]
}
func (ml middlewares) Less(i, j int) bool {
	return ml[i].pos < ml[j].pos
}
