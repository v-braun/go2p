package middleware

import (
	"fmt"

	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/core/logging"
)

var annotationKey = "middleware.routes"

// RoutingTable represents handler registered by a path.
// A message will be checked for the existence of an annotation with the name "__routes_path"
// and this value will be used to find a route within the routing table
type RoutingTable *map[string]func(peer *core.Peer, msg *core.Message)

// EmptyRoutesTable is a table without any routes
var EmptyRoutesTable = *new(RoutingTable)

// Routes provides an route based middleware
// You can listen to specific endpoints and send messages to them
// This is similar to a controller/action pattern in HTTP frameworks
func Routes(rt RoutingTable) (string, core.MiddlewareFunc) {
	if rt == EmptyRoutesTable {
		return "routes", func(peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
			return core.Next, nil
		}
	}

	f := func(peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
		op, err := middlewareRoutesImpl(rt, peer, pipe, msg)
		return op, err
	}
	return "routes", f
}

// NewMessageRoutedFromString creates a new routed message to the handler given by path
// with the provided string content
func NewMessageRoutedFromString(path string, content string) *core.Message {
	msg := NewMessageRoutedFromData(path, []byte(content))
	return msg
}

// NewMessageRoutedFromData creates a new routed message to the handler given by path
// with the provided data
func NewMessageRoutedFromData(path string, data []byte) *core.Message {
	msg := core.NewMessageFromData(data)
	msg.Metadata().Put(annotationKey, path)
	return msg
}

func middlewareRoutesImpl(rt RoutingTable, peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
	var log = logging.NewLogger("middleware_routes")
	if pipe.Operation() == core.Send {
		return core.Next, nil
	}

	routeHdr, found := msg.Metadata().Get(annotationKey)
	if !found {
		log.Debug(logging.Fields{}, fmt.Sprintf("msg has no %s key, skip routing", annotationKey))
		return core.Next, nil
	}

	routeStr := routeHdr.(string)
	route, hasRoute := (*rt)[routeStr]
	if !hasRoute {
		log.Warning(logging.Fields{
			"route": route,
			"table": rt,
		}, "found routing key in message, but miss value in routing table")
		return core.Next, nil
	}

	log.Debug(logging.Fields{"route": routeStr}, "execute route")
	go route(peer, msg)

	return core.Next, nil
}
