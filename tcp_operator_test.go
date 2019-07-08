package go2p

import (
	"fmt"
	"sync"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
)

func TestTCPOperatorNegativeCases(t *testing.T) {
	op := NewTCPOperator("ttt", "10.10.10.10")

	err := op.Dial("ttt", "foo")
	assert.Error(t, err)

	err = op.Dial("tcp", "foo")
	assert.Error(t, err)

	err = op.Start()
	assert.Error(t, err)

	op = NewTCPOperator("tcp", "foo")
	err = op.Start()
	assert.Error(t, err)

	port, _ := freeport.GetFreePort()
	op = NewTCPOperator("tcp", fmt.Sprintf("localhost:%d", port))
	onErrCalled := new(sync.WaitGroup)
	onErrCalled.Add(1)
	op.OnError(func(err error) {
		assert.Error(t, err)
		onErrCalled.Done()
	})

	op.Start()
	op.server.Close()

	onErrCalled.Wait()

}

func TestPingPong(t *testing.T) {
	clientsWg := new(sync.WaitGroup)
	clientsWg.Add(2)

	msgWg := new(sync.WaitGroup)
	msgWg.Add(2)

	op1 := NewTCPOperator("tcp", "127.0.0.1:3377")
	op2 := NewTCPOperator("tcp", "127.0.0.1:3378")

	conn1 := NewNetworkConnection().
		WithOperator(op1).
		Build()

	conn2 := NewNetworkConnection().
		WithOperator(op2).
		Build()

	conn1.OnPeer(func(p *Peer) {
		clientsWg.Done()

		if p.RemoteAddress() != "tcp:127.0.0.1:3378" {
			t.Fatal("unexpected address", p.RemoteAddress())
			return
		}

		clientsWg.Wait()
		conn1.Send(NewMessageFromData([]byte("hello")), p.RemoteAddress())
	})

	conn2.OnPeer(func(p *Peer) {
		clientsWg.Done()
	})

	conn1.OnMessage(func(p *Peer, m *Message) {
		assert.Equal(t, "hello back", m.PayloadGetString())
		fmt.Printf("from %s: %s\n", p.RemoteAddress(), m.PayloadGetString())
		msgWg.Done()
	})

	conn2.OnMessage(func(p *Peer, m *Message) {
		assert.Equal(t, "hello", m.PayloadGetString())
		fmt.Printf("from %s: %s\n", p.RemoteAddress(), m.PayloadGetString())
		go conn2.Send(NewMessageFromData([]byte("hello back")), p.RemoteAddress())
		msgWg.Done()
	})

	err := conn1.Start()
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, err)

	err = conn2.Start()
	if err != nil {
		t.Fatal(err)
	}

	conn1.ConnectTo("tcp", "localhost:3378")

	clientsWg.Wait()
	msgWg.Wait()

	conn1.Stop()
	conn2.Stop()
}
