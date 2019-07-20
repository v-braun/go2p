package go2p_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/v-braun/go2p"
)

func getPorts(t *testing.T, amount int) []int {
	var result []int
	for i := 0; i < amount; i++ {
		p, err := freeport.GetFreePort()
		assert.NoError(t, err)
		result = append(result, p)
	}

	return result
}

func TestPingPong(t *testing.T) {
	type sut struct {
		addr        string
		net         *go2p.NetworkConnection
		store       go2p.PeerStore
		connections int
	}

	var suts []*sut
	ports := getPorts(t, 3)
	for _, p := range ports {
		s := new(sut)
		s.addr = fmt.Sprintf("127.0.0.1:%d", p)
		s.store = go2p.NewDefaultPeerStore(1, 1)
		op := go2p.NewTCPOperator("tcp", s.addr)

		s.net = go2p.NewNetworkConnection().
			WithOperator(op).
			WithPeerStore(s.store).
			Build()

		suts = append(suts, s)
	}

	wgTestDone := new(sync.WaitGroup)
	wgTestDone.Add(1)

	peerDisconnected := false
	suts[0].net.OnPeer(func(p *go2p.Peer) {
		suts[2].net.ConnectTo("tcp", suts[0].addr)
	})
	suts[0].net.OnPeerDisconnect(func(p *go2p.Peer) {
		peerDisconnected = true
	})

	suts[0].store.OnPeerWantRemove(func(p *go2p.Peer) {
		wgTestDone.Done()
	})

	suts[0].net.Start()
	suts[1].net.Start()
	suts[2].net.Start()

	suts[1].net.ConnectTo("tcp", suts[0].addr)

	wgTestDone.Wait()

	suts[0].net.Stop()
	suts[1].net.Stop()
	suts[2].net.Stop()

	assert.True(t, peerDisconnected)
}
