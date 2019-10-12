package core_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/internal"
)

func TestConnect(t *testing.T) {
	sut1 := internal.CreateTCPSUT(t)
	sut2 := internal.CreateTCPSUT(t)

	sut1.Wg.Add(1)
	sut2.Wg.Add(1)

	sut1.OnPeer(func(p *core.Peer) {
		sut1.Wg.Done()
	})
	sut2.OnPeer(func(p *core.Peer) {
		sut2.Wg.Done()
	})

	sut1.Start()
	defer sut1.Stop()

	sut2.Start()
	defer sut2.Stop()

	sut1.ConnectTo("tcp", sut2.Addr)

	sut1.Wg.Wait()
	sut2.Wg.Wait()

}

func TestDisconnectOnStop(t *testing.T) {
	sut1 := internal.CreateTCPSUT(t)
	sut2 := internal.CreateTCPSUT(t)

	sut1.Wg.Add(1)
	sut2.Wg.Add(1)

	sut1.OnPeer(func(p *core.Peer) {
		sut1.Wg.Done()
	})

	sut2.OnPeerDisconnect(func(p *core.Peer) {
		sut2.Wg.Done()
	})

	sut1.Start()
	sut2.Start()

	sut1.ConnectTo("tcp", sut2.Addr)
	sut1.Wg.Wait() // wait for connect event
	sut1.Stop()

	sut2.Wg.Wait() // wait for disconnect event

	sut2.Stop()
}

func TestDisconnect(t *testing.T) {
	sut1 := internal.CreateTCPSUT(t)
	sut2 := internal.CreateTCPSUT(t)

	sut1.Wg.Add(1)
	sut2.Wg.Add(1)

	sut1.OnPeer(func(p *core.Peer) {
		sut1.Wg.Done()
	})

	sut2.OnPeerDisconnect(func(p *core.Peer) {
		sut2.Wg.Done()
	})

	sut1.Start()
	sut2.Start()

	sut1.ConnectTo("tcp", sut2.Addr)
	sut1.Wg.Wait() // wait for connect event
	sut1.DisconnectFrom(sut2.FullAddr)
	time.Sleep(time.Second * 1)
	sut2.Wg.Wait() // wait for disconnect event

	sut1.Stop()
	sut2.Stop()
}

func TestSend(t *testing.T) {
	sut1 := internal.CreateTCPSUT(t)
	sut2 := internal.CreateTCPSUT(t)

	sut1.Wg.Add(1)

	expectedMessage := "hello world"
	receivedMessage := ""

	sut1.OnMessage(func(p *core.Peer, m *core.Message) {
		receivedMessage = m.PayloadGetString()
		sut1.Wg.Done()
	})

	sut1.Start()
	defer sut1.Stop()
	sut2.Start()
	defer sut2.Stop()

	sut2.ConnectTo("tcp", sut1.Addr)
	sut2.Send(core.NewMessageFromString(expectedMessage), sut1.FullAddr)

	sut1.Wg.Wait()

	assert.Equal(t, expectedMessage, receivedMessage)
}

func TestRouteEnabled(t *testing.T) {
	sut1 := internal.CreateTCPSUT(t)
	sut2 := internal.CreateTCPSUT(t)

	disabledCalled := false

	sut1.PrependMiddleware(core.NewMiddleware("disabled", func(peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
		disabledCalled = true
		return core.Next, nil
	}))
	sut1.PrependMiddleware(core.NewMiddleware("enabled", func(peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
		sut1.Wg.Done()
		return core.Next, nil
	}))

	sut1.Start()
	defer sut1.Stop()
	sut2.Start()
	defer sut2.Stop()

	sut2.ConnectTo("tcp", sut1.Addr)

	sut1.Wg.Add(1)
	sut1.SetMiddlewareEnabled("disabled", false)
	sut2.Send(core.NewMessageFromString("#1"), sut1.FullAddr)
	sut1.Wg.Wait()
	assert.False(t, disabledCalled)

	sut1.Wg.Add(1)
	sut1.SetMiddlewareEnabled("disabled", true)
	sut2.Send(core.NewMessageFromString("#2"), sut1.FullAddr)
	sut1.Wg.Wait()
	assert.True(t, disabledCalled)
}

func TestBroadcast(t *testing.T) {
	sut1 := internal.CreateTCPSUT(t)
	sut2 := internal.CreateTCPSUT(t)
	sut3 := internal.CreateTCPSUT(t)

	sut1.Wg.Add(1)
	sut2.Wg.Add(1)
	sut3.Wg.Add(1)

	expectedMessage := "hello world"
	receivedMessage := ""

	sut1.OnMessage(func(p *core.Peer, m *core.Message) {
		receivedMessage = m.PayloadGetString()
		sut1.Wg.Done()
	})
	sut2.OnMessage(func(p *core.Peer, m *core.Message) {
		receivedMessage = m.PayloadGetString()
		sut2.Wg.Done()
	})

	sut1.Start()
	defer sut1.Stop()
	sut2.Start()
	defer sut2.Stop()
	sut3.Start()
	defer sut3.Stop()

	sut3.ConnectTo("tcp", sut1.Addr)
	sut3.ConnectTo("tcp", sut2.Addr)
	msg := core.NewMessageFromString("")
	msg.PayloadSetString(expectedMessage)
	sut3.SendBroadcast(msg)

	sut1.Wg.Wait()
	sut2.Wg.Wait()

	assert.Equal(t, expectedMessage, receivedMessage)
}

func TestInvalidAddress(t *testing.T) {
	sut1 := internal.CreateTCPSUT(t)

	sut1.Start()
	err := sut1.ConnectTo("tcp", "foo")
	assert.NotEqual(t, core.ErrInvalidNetwork, err)
	assert.Error(t, err)
}

func TestOperatorStartError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sut1 := internal.CreateMockSUT(t, ctrl)
	sut1.ExpectedErr = errors.New("expected err")

	sut1.Operator.EXPECT().Start().Return(sut1.ExpectedErr)
	sut1.ActualErr = sut1.Start()

	assert.Equal(t, sut1.ExpectedErr, sut1.ActualErr)
}

func TestConnReadErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sut1 := internal.CreateMockSUT(t, ctrl)
	sut1.ExpectedErr = errors.New("expected err")

	sut1.Operator.EXPECT().Start().Return(nil)

	sut1.Conn.EXPECT().ReadMessage().DoAndReturn(func() (*core.Message, error) {
		return nil, sut1.ExpectedErr
	})
	sut1.Conn.EXPECT().Close()

	onDisconn := sut1.AwaitPeerDisconnect()
	onErr := sut1.AwaitPeerError()

	sut1.Start()

	sut1.NotifyNewConn(sut1.Conn)
	sut1.Wg.Wait()

	sut1.ActualErr = <-onErr
	<-onDisconn

	assert.Equal(t, sut1.ExpectedErr, errors.Cause(sut1.ActualErr))
}

func TestConnReadDisconnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sut1 := internal.CreateMockSUT(t, ctrl)

	sut1.Operator.EXPECT().Start().Return(nil)

	notifyReadDisconn := sut1.CreateDisconnectOnReadMessage()
	sut1.Conn.EXPECT().Close()

	onDisconn := sut1.AwaitPeerDisconnect()

	sut1.Start()

	sut1.NotifyNewConn(sut1.Conn)

	close(notifyReadDisconn)
	<-onDisconn
}

func TestOperatorConnectError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sut1 := internal.CreateMockSUT(t, ctrl)
	sut1.Operator.EXPECT().Dial(gomock.Any(), gomock.Any()).Return(sut1.ExpectedErr)
	sut1.Operator.EXPECT().Start().Return(nil)
	sut1.Operator.EXPECT().Stop()

	sut1.Start()

	sut1.ActualErr = sut1.ConnectTo("any", "any")

	sut1.Stop()

	assert.Equal(t, sut1.ExpectedErr, sut1.ActualErr)
}

func TestEnsureStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sut1 := internal.CreateMockSUT(t, ctrl)
	sut1.Operator.EXPECT().Start().Return(nil)
	sut1.Operator.EXPECT().Stop()

	assert.Panics(t, func() {
		sut1.ConnectTo("any", "any")
	})

	sut1.Start()
	defer sut1.Stop()

	assert.Panics(t, func() {
		sut1.OnPeerDisconnect(func(p *core.Peer) {

		})
	})

}

func TestConnWriteErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sut1 := internal.CreateMockSUT(t, ctrl)
	sut1.ExpectedErr = errors.New("expected err")

	sut1.Operator.EXPECT().Start().Return(nil)

	notifyReadDisconn := sut1.CreateDisconnectOnReadMessage()

	sut1.Conn.EXPECT().WriteMessage(gomock.Any()).Return(sut1.ExpectedErr)
	sut1.Conn.EXPECT().Close()

	onDisconn := sut1.AwaitPeerDisconnect()
	onErr := sut1.AwaitPeerError()

	sut1.Start()

	sut1.NotifyNewConn(sut1.Conn)

	sut1.Network.Send(core.NewMessageFromString("hello"), "remote")

	sut1.ActualErr = <-onErr
	close(notifyReadDisconn) // unblock ReadMessage
	<-onDisconn

	assert.Equal(t, sut1.ExpectedErr, errors.Cause(sut1.ActualErr))
}
