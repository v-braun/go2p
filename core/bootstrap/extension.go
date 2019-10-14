package bootstrap

import (
	"path/filepath"

	"github.com/v-braun/go-must"
	"github.com/v-braun/go2p/core"
)

var _ core.Extension = (*extension)(nil)

type extension struct {
	confFilePath string
}

func NewBootstrapper(confPath string) core.Extension {
	confPath, err := filepath.Abs(confPath)
	must.NoError(err, "unexpected error for path "+confPath)
	result := &extension{
		confFilePath: confPath,
	}

	// if _, err := os.Stat(result.confFilePath); os.IsNotExist(err) {
	// 	os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0666)
	// }

	return result
}

func (ex *extension) Install(n *core.Network) error {
	go func() {
		<-n.Started()
	}()
	return nil
}

func (ex *extension) Uninstall() {

}

func (ex *extension) BeforePeerConnect(p *core.Peer) error {
	return nil
}

func (ex *extension) AfterPeerConnect(p *core.Peer) {

}
