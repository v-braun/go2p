package go2p

import (
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/middleware"
	"github.com/v-braun/go2p/tcp"
)

func NewBareNetwork() *core.Network {
	return core.NewNetwork()
}


func NewTcpNetwork(localAddr string) *core.Network {
	net := core.NewNetwork().
		UseOperator(tcp.NewOperator("tcp", localAddr)).
		AppendMiddleware(middleware.Headers()).
		AppendMiddleware(middleware.Crypt()).
		AppendMiddleware(middleware.Log())

	return net
}
