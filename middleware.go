package go2p

import (
	"fmt"
	"sort"
)

// MiddlewareResult represents a result returned by a middleware
// possible values are *Stop* and *Next*
type MiddlewareResult int

const (
	// Stop will be returned by a middleware when the pipe execution should be stopped
	Stop MiddlewareResult = iota

	// Next will be returned by a middleware when the pipe execution should be continued
	Next MiddlewareResult = iota
)

// MiddlewareFunc represents a middleware implementation function
type MiddlewareFunc func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error)

// Middleware represents a wrapped middleware function with
// additional information for internal usage
type Middleware struct {
	execute MiddlewareFunc
	name    string
	pos     int
}

// NewMiddleware wraps the provided action into a Middleware instance
func NewMiddleware(name string, action MiddlewareFunc) *Middleware {
	return &Middleware{
		name:    name,
		execute: action,
	}
}

// String returns the string representation of this instance
func (m *Middleware) String() string {
	return fmt.Sprintf("%s (%d)", m.name, m.pos)
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
