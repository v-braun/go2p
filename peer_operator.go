package go2p

import "errors"

var InvalidNetworkError = errors.New("invalid network")

type PeerOperator interface {
	Dial(network string, addr string) error
	OnPeer(handler func(p Adapter))
	Start() error
	Stop()
}
