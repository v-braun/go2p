package middleware

import (
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/core/logging"
)

// Log creates a logging middleware for in and outgoing messages
func Log() *core.Middleware {
	return core.NewMiddleware("log", middlewareLogImpl)
}

func middlewareLogImpl(peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
	directions := make(map[core.PipeOperation]string)
	directions[core.Send] = "out->"
	directions[core.Receive] = "<--in"

	logging.NewLogger("middleware_log").Debug(logging.Fields{
		"remote": peer.RemoteAddress(),
		"local":  peer.LocalAddress(),
		"len":    len(msg.PayloadGet()),
	}, "message "+directions[core.Send])

	return core.Next, nil
}
