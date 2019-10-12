package internal

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/phayes/freeport"
	"github.com/v-braun/go2p"
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/core/mocks"
)

type TCPSUT struct {
	*core.Network
	Addr     string
	FullAddr string
	Wg       sync.WaitGroup
}

type MockSUT struct {
	*core.Network

	Operator *mocks.MockOperator
	Conn     *mocks.MockConn
	Wg       sync.WaitGroup

	ExpectedErr error
	ActualErr   error

	NotifyNewConn func(p core.Conn)
}

func (m *MockSUT) CreateDisconnectOnReadMessage() chan struct{} {
	done := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(1)
	m.Conn.EXPECT().ReadMessage().DoAndReturn(func() (*core.Message, error) {
		wg.Wait()
		return nil, core.DisconnectedError
	})

	go func() {
		select {
		case <-done:
			wg.Done()
		}
	}()

	return done
}

func (m *MockSUT) AwaitPeerError() chan error {
	done := make(chan error)

	m.OnPeerError(func(p *core.Peer, err error) {
		done <- err
	})

	return done

}

func (m *MockSUT) AwaitPeerDisconnect() chan struct{} {
	done := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(1)
	m.OnPeerDisconnect(func(p *core.Peer) {
		wg.Done()
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	return done

}

func CreateTCPSUT(t *testing.T) *TCPSUT {
	p1, err := freeport.GetFreePort()
	assert.NoError(t, err)

	addr := fmt.Sprintf("127.0.0.1:%d", p1)
	full := "tcp:" + addr
	net := go2p.NewTcpNetwork(addr)

	return &TCPSUT{Network: net, Addr: addr, FullAddr: full, Wg: sync.WaitGroup{}}
}

func CreateMockSUT(t *testing.T, controller *gomock.Controller) *MockSUT {
	conn := mocks.NewMockConn(controller)
	op := mocks.NewMockOperator(controller)
	net := go2p.NewBareNetwork()
	net.UseOperator(op)

	result := &MockSUT{Network: net, Conn: conn, Operator: op, Wg: sync.WaitGroup{}}
	result.ExpectedErr = errors.New("expected")
	op.EXPECT().OnPeer(gomock.Any()).Do(func(handler func(p core.Conn)) {
		result.NotifyNewConn = handler
	}).AnyTimes()

	conn.EXPECT().RemoteAddress().AnyTimes().Return("remote")
	conn.EXPECT().LocalAddress().AnyTimes().Return("local")

	// TODO: should also register an onerror handler
	//sut1.operator.EXPECT().OnError(gomock.Any())

	return result
}
