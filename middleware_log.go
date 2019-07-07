package go2p

import (
	"fmt"
)

func Log() (string, MiddlewareFunc) {
	return "log", middlewareLogImpl
}

func middlewareLogImpl(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
	directions := make(map[PipeOperation]string)
	directions[Send] = "out->"
	directions[Receive] = "<--in"

	txt := fmt.Sprintf("%s %s (%d bytes)", peer.Address(), directions[pipe.Operation()], len(msg.PayloadGet()))
	fmt.Sprintln(txt)

	return Next, nil
}
