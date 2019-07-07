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
		action.pos = (idx + 1)
		result = append(result, action)
	}

	return result
}

func (ml middlewares) nextItems(op PipeOperation, pos int) middlewares {
	result := ml.Copy()
	sort.Sort(result)
	result = result[pos:]
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

func (ml middlewares) Min() int {
	if len(ml) <= 0 {
		panic("list is empty")
	}

	min := ml[0].pos
	for _, item := range ml {
		if item.pos < min {
			min = item.pos
		}
	}

	return min
}

func (ml middlewares) Max() int {
	if len(ml) <= 0 {
		panic("list is empty")
	}

	max := ml[0].pos

	for _, item := range ml {
		if item.pos > max {
			max = item.pos
		}
	}

	return max
}

func (ml middlewares) ByPos(pos int) *Middleware {
	for _, item := range ml {
		if item.pos == pos {
			return item
		}
	}

	return nil
}

func (ml middlewares) String() string {
	result := ""
	for _, item := range ml {
		result += fmt.Sprintf("%s(%d) -> ", item.name, item.pos)
	}

	return result
}
