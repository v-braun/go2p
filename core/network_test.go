package core_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/phayes/freeport"
	"github.com/v-braun/go2p"
	"github.com/v-braun/go2p/core"
)

type tcpSUT struct {
	*core.Network
	addr     string
	fullAddr string
	wg       sync.WaitGroup
}

type mockSUT struct {
	*core.Network
}

func createTCPSUT(t *testing.T) *tcpSUT {
	p1, err := freeport.GetFreePort()
	assert.NoError(t, err)

	addr := fmt.Sprintf("127.0.0.1:%d", p1)
	full := "tcp:" + addr
	net := go2p.NewTcpNetwork(addr)

	return &tcpSUT{Network: net, addr: addr, fullAddr: full, wg: sync.WaitGroup{}}
}

// func createMockSUT(t *testing.T, controller *gomock.Controller) *mockSUT {
// 	conn := mock_core.NewMockConn()
// 	op := mock_core.NewMockOperator()
// 	result := &mockSUT{conn: conn, operator: op}

// 	// conn := mock_core.New
// 	return result
// }

func TestConnect(t *testing.T) {
	sut1 := createTCPSUT(t)
	sut2 := createTCPSUT(t)

	sut1.wg.Add(1)
	sut2.wg.Add(1)

	sut1.OnPeer(func(p *core.Peer) {
		sut1.wg.Done()
	})
	sut2.OnPeer(func(p *core.Peer) {
		sut2.wg.Done()
	})

	sut1.Start()
	sut2.Start()

	sut1.ConnectTo("tcp", sut2.addr)

	sut1.wg.Wait()
	sut2.wg.Wait()

	sut1.Stop()
	sut2.Stop()
}

func TestDisconnectOnStop(t *testing.T) {
	sut1 := createTCPSUT(t)
	sut2 := createTCPSUT(t)

	sut1.wg.Add(1)
	sut2.wg.Add(1)

	sut1.OnPeer(func(p *core.Peer) {
		sut1.wg.Done()
	})

	sut2.OnPeerDisconnect(func(p *core.Peer) {
		sut2.wg.Done()
	})

	sut1.Start()
	sut2.Start()

	sut1.ConnectTo("tcp", sut2.addr)
	sut1.wg.Wait() // wait for connect event
	sut1.Stop()

	sut2.wg.Wait() // wait for disconnect event

	sut2.Stop()
}

func TestDisconnect(t *testing.T) {
	sut1 := createTCPSUT(t)
	sut2 := createTCPSUT(t)

	sut1.wg.Add(1)
	sut2.wg.Add(1)

	sut1.OnPeer(func(p *core.Peer) {
		sut1.wg.Done()
	})

	sut2.OnPeerDisconnect(func(p *core.Peer) {
		sut2.wg.Done()
	})

	sut1.Start()
	sut2.Start()

	sut1.ConnectTo("tcp", sut2.addr)
	sut1.wg.Wait() // wait for connect event
	sut1.DisconnectFrom(sut2.fullAddr)

	sut2.wg.Wait() // wait for disconnect event

	sut1.Stop()
	sut2.Stop()
}

func TestSend(t *testing.T) {
	sut1 := createTCPSUT(t)
	sut2 := createTCPSUT(t)

	sut1.wg.Add(1)

	expectedMessage := "hello world"
	receivedMessage := ""

	sut1.OnMessage(func(p *core.Peer, m *core.Message) {
		receivedMessage = m.PayloadGetString()
		sut1.wg.Done()
	})

	sut1.Start()
	sut2.Start()

	sut2.ConnectTo("tcp", sut1.addr)
	sut2.Send(core.NewMessageFromString(expectedMessage), sut1.fullAddr)

	sut1.wg.Wait()

	assert.Equal(t, expectedMessage, receivedMessage)

	sut1.Stop()
	sut2.Stop()
}

func TestBroadcast(t *testing.T) {
	sut1 := createTCPSUT(t)
	sut2 := createTCPSUT(t)
	sut3 := createTCPSUT(t)

	sut1.wg.Add(1)
	sut2.wg.Add(1)
	sut3.wg.Add(1)

	expectedMessage := "hello world"
	receivedMessage := ""

	sut1.OnMessage(func(p *core.Peer, m *core.Message) {
		receivedMessage = m.PayloadGetString()
		sut1.wg.Done()
	})
	sut2.OnMessage(func(p *core.Peer, m *core.Message) {
		receivedMessage = m.PayloadGetString()
		sut2.wg.Done()
	})

	sut1.Start()
	sut2.Start()
	sut3.Start()

	sut3.ConnectTo("tcp", sut1.addr)
	sut3.ConnectTo("tcp", sut2.addr)
	msg := core.NewMessageFromString("")
	msg.PayloadSetString(expectedMessage)
	sut3.SendBroadcast(msg)

	sut1.wg.Wait()
	sut2.wg.Wait()

	assert.Equal(t, expectedMessage, receivedMessage)

	sut1.Stop()
	sut2.Stop()
	sut3.Stop()
}

func TestInvalidAddress(t *testing.T) {
	sut1 := createTCPSUT(t)

	sut1.Start()
	err := sut1.ConnectTo("tcp", "foo")
	assert.NotEqual(t, core.ErrInvalidNetwork, err)
	assert.Error(t, err)
}

func TestOperatorStartError(t *testing.T) {

}
