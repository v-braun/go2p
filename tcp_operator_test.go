package go2p_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/v-braun/go2p"
)

func TestPingPong(t *testing.T) {
	clientsWg := new(sync.WaitGroup)
	clientsWg.Add(2)

	msgWg := new(sync.WaitGroup)
	msgWg.Add(2)

	op1 := go2p.NewTcpOperator("tcp", "127.0.0.1:3377")
	op2 := go2p.NewTcpOperator("tcp", "127.0.0.1:3378")

	conn1 := go2p.NewNetworkConnection().
		WithOperator(op1).
		Build()

	conn2 := go2p.NewNetworkConnection().
		WithOperator(op2).
		Build()

	conn1.OnPeer(func(p *go2p.Peer) {
		clientsWg.Done()

		if p.RemoteAddress() != "tcp:127.0.0.1:3378" {
			t.Fatal("unexpected address", p.RemoteAddress())
			return
		}

		clientsWg.Wait()
		conn1.Send(go2p.NewMessageFromData([]byte("hello")), p.RemoteAddress())
	})

	conn2.OnPeer(func(p *go2p.Peer) {
		clientsWg.Done()
	})

	conn1.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		assert.Equal(t, "hello back", m.PayloadGetString())
		fmt.Printf("from %s: %s\n", p.RemoteAddress(), m.PayloadGetString())
		msgWg.Done()
	})

	conn2.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		assert.Equal(t, "hello", m.PayloadGetString())
		fmt.Printf("from %s: %s\n", p.RemoteAddress(), m.PayloadGetString())
		go conn2.Send(go2p.NewMessageFromData([]byte("hello back")), p.RemoteAddress())
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
