package core

import (
	"fmt"
	"sort"
)

type MiddlewareResult int

const (
	Stop MiddlewareResult = iota
	Next MiddlewareResult = iota
)

type MiddlewareFunc func(pipe Pipe, msg Message) (MiddlewareResult, error)

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

func newMiddlewares(actions) middlewares {
	res := middlewareList{}
	for i := 0; i < amount; i++ {
		m := createMiddleware(i)
		res = append(res, m)
	}

	for i, m := range ml {
		m.pos = i + 1
	}
}

func (ml middlewareList) nextItems(op PipeOperation, pos int) middlewareList {
	result := ml.Copy()
	sort.Sort(result)
	result = result[pos:]
	if op == Receive {
		sort.Sort(sort.Reverse(result))
	}

	return result
}

func (ml middlewareList) Copy() middlewareList {
	result := make(middlewareList, len(ml))
	copy(result, ml)
	return result
}

func (ml middlewareList) Len() int {
	return len(ml)
}

func (ml middlewareList) Swap(i, j int) {
	ml[i], ml[j] = ml[j], ml[i]
}
func (ml middlewareList) Less(i, j int) bool {
	return ml[i].pos < ml[j].pos
}

func (ml middlewareList) Min() int {
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

func (ml middlewareList) Max() int {
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

func (ml middlewareList) ByPos(pos int) *Middleware {
	for _, item := range ml {
		if item.pos == pos {
			return item
		}
	}

	return nil
}
