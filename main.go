package go2p

import (
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/middleware"
	"github.com/v-braun/go2p/mockcore"
	"github.com/v-braun/go2p/tcp"
)

type test struct {
	operator *mockcore.MockOperator
}

func NewBareNetwork() *core.Network {
	return core.NewNetwork()
}

func NewTcpNetwork(localAddr string) *core.Network {
	net := core.NewNetwork().
		UseOperator(tcp.NewOperator("tcp", localAddr)).
		UseMiddleware(middleware.Headers()).
		UseMiddleware(middleware.Crypt()).
		UseMiddleware(middleware.Log())

	return net
}
