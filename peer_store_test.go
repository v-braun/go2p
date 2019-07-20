package go2p

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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

func TestErrHandlingOnInvalidAdd(t *testing.T) {
	store := new(MockPeerStore)
	store.On("AddPeer", mock.Anything).Return(errors.New(""))
	store.On("OnPeerAdd", mock.Anything)
	store.On("OnPeerWantRemove", mock.Anything)
	store.On("Start", mock.Anything)
	store.On("Stop", mock.Anything)

	ports := getPorts(t, 2)
	addr1 := fmt.Sprintf("127.0.0.1:%d", ports[0])
	op1 := NewTCPOperator("tcp", addr1)

	addr2 := fmt.Sprintf("127.0.0.1:%d", ports[1])
	op2 := NewTCPOperator("tcp", addr2)

	net1 := NewNetworkConnection().
		WithOperator(op1).
		WithPeerStore(store).
		Build()

	net2 := NewNetworkConnection().
		WithOperator(op2).
		WithPeerStore(store).
		Build()

	net1.Start()
	net2.Start()

	net1.ConnectTo("tcp", addr2)

}

func TestPeerErrorHandling(t *testing.T) {
	store1 := NewDefaultPeerStore(2, 2)
	store2 := NewDefaultPeerStore(2, 2)

	ports := getPorts(t, 2)
	addr1 := fmt.Sprintf("127.0.0.1:%d", ports[0])
	op1 := NewTCPOperator("tcp", addr1)

	addr2 := fmt.Sprintf("127.0.0.1:%d", ports[1])
	op2 := NewTCPOperator("tcp", addr2)

	net1 := NewNetworkConnection().
		WithOperator(op1).
		WithPeerStore(store1).
		Build()

	net2 := NewNetworkConnection().
		WithOperator(op2).
		WithPeerStore(store2).
		Build()

	net2.OnPeer(func(p *Peer) {
		p.emitter.EmitAsync("error", p, errors.New("fail"))
	})

	wgTestDone := new(sync.WaitGroup)
	wgTestDone.Add(1)
	net2.OnPeerError(func(p *Peer, err error) {
		wgTestDone.Done()
	})

	net1.Start()
	net2.Start()

	net1.ConnectTo("tcp", addr2)

	wgTestDone.Wait()

}

func TestCapLimits(t *testing.T) {
	type sut struct {
		addr        string
		net         *NetworkConnection
		store       PeerStore
		connections int
	}

	var suts []*sut
	ports := getPorts(t, 3)
	for _, p := range ports {
		s := new(sut)
		s.addr = fmt.Sprintf("127.0.0.1:%d", p)
		s.store = NewDefaultPeerStore(1, 1)
		op := NewTCPOperator("tcp", s.addr)

		s.net = NewNetworkConnection().
			WithOperator(op).
			WithPeerStore(s.store).
			Build()

		suts = append(suts, s)
	}

	wgTestDone := new(sync.WaitGroup)
	wgTestDone.Add(1)

	peerDisconnected := false
	suts[0].net.OnPeer(func(p *Peer) {
		suts[2].net.ConnectTo("tcp", suts[0].addr)
	})
	suts[0].net.OnPeerDisconnect(func(p *Peer) {
		peerDisconnected = true
	})

	suts[0].store.OnPeerWantRemove(func(p *Peer) {
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
