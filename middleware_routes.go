package go2p

var annotationKey = "middleware.routes"

// RoutingTable represents handler registered by a path.
// A message will be checked for the existence of an annotation with the name "__routes_path"
// and this value will be used to find a route within the routing table
type RoutingTable *map[string]func(peer *Peer)

// EmptyRoutesTable is a table without any routes
var EmptyRoutesTable = *new(RoutingTable)

// Routes provides an route based middleware
// You can listen to specific endpoints and send messages to them
// This is similar to a controller/action pattern in HTTP frameworks
func Routes(rt RoutingTable) (string, MiddlewareFunc) {
	if rt == EmptyRoutesTable {
		return "routes", func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
			return Next, nil
		}
	}

	f := func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
		op, err := middlewareRoutesImpl(rt, peer, pipe, msg)
		return op, err
	}
	return "routes", f
}

// NewMessageRoutedFromString creates a new routed message to the handler given by path
// with the provided string content
func NewMessageRoutedFromString(path string, content string) *Message {
	msg := NewMessageFromString(content)
	msg.Metadata().Put(annotationKey, path)
	return msg
}

// NewMessageRoutedFromString creates a new routed message to the handler given by path
// with the provided data
func NewMessageRoutedFromData(path string, data []byte) *Message {
	msg := NewMessageFromData(data)
	msg.Metadata().Put(annotationKey, path)
	return msg
}

func middlewareRoutesImpl(rt RoutingTable, peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
	routeHdr, found := msg.Metadata().Get(annotationKey)
	if !found {
		return Next, nil
	}

	routeStr := routeHdr.(string)
	route, hasRoute := (*rt)[routeStr]
	if !hasRoute {
		return Next, nil
	}

	route(peer)

	return Next, nil
}
