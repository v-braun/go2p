package extensions

import "github.com/v-braun/go2p/core"

var _ core.Extension = (*bootstrap)(nil)

type bootstrap struct {
}

func (b *bootstrap) Install(n *core.Network) error {
	return nil
}

func (b *bootstrap) Uninstall() {

}

func (b *bootstrap) BeforePeerConnect(p *core.Peer) error {
	return nil
}

func (b *bootstrap) AfterPeerConnect(p *core.Peer) {

}
