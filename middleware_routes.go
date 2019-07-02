package go2p

var annotationKey = "__routes_path"

type RoutingTable map[string]func(peer *Peer)

func Routes(rt RoutingTable) (string, MiddlewareFunc) {

	f := func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
		op, err := middlewareRoutesImpl(rt, peer, pipe, msg)
		return op, err
	}
	return "routes", f
}

func NewMessageRoutedFromString(path string, content string) *Message {
	msg := NewMessageFromString(content)
	msg.Metadata().Put(annotationKey, path)
	return msg
}

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
	route, hasRoute := rt[routeStr]
	if !hasRoute {
		return Next, nil
	}

	route(peer)

	return Next, nil
}
