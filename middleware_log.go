package go2p

import (
	"fmt"
)

// Log creates a logging middleware for in and outgoing messages
func Log() (string, MiddlewareFunc) {
	return "log", middlewareLogImpl
}

func middlewareLogImpl(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
	directions := make(map[PipeOperation]string)
	directions[Send] = "out->"
	directions[Receive] = "<--in"

	txt := fmt.Sprintf("%s %s (%d bytes) - local endpoint: %s", peer.RemoteAddress(), directions[pipe.Operation()], len(msg.PayloadGet()), peer.LocalAddress())
	newLogger("middleware_log").Debug(txt)

	return Next, nil
}
