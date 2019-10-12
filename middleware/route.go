package middleware

import (
	"fmt"

	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/core/logging"
)

func Route(route string, handler func(peer *core.Peer, msg *core.Message)) *core.Middleware {
	return core.NewMiddleware(GetMiddlewareNameFromRoute(route), func(peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
		return middlewareRouteImpl(peer, pipe, msg, route, handler)
	})
}

func GetMiddlewareNameFromRoute(route string) string {
	return "route-" + route
}

func AssignRouteToMessage(route string, msg *core.Message) *core.Message {
	var annotationKey = "m.r." + route
	msg.Metadata().Put(annotationKey, true)
	return msg
}

func middlewareRouteImpl(peer *core.Peer, pipe *core.Pipe, msg *core.Message, route string, handler func(peer *core.Peer, msg *core.Message)) (core.MiddlewareResult, error) {
	var annotationKey = "m.r." + route
	var log = logging.NewLogger("middleware_route_" + route)
	if pipe.Operation() == core.Send {
		k, ok := msg.Metadata().Get(annotationKey)
		fmt.Println("send->", k, ok)
		return core.Next, nil
	}

	_, found := msg.Metadata().Get(annotationKey)
	if found {
		log.Debug(logging.Fields{"match": true}, "route match")
		go handler(peer, msg)
	} else {
		log.Debug(logging.Fields{"match": false}, "route match")
	}

	return core.Next, nil
}
