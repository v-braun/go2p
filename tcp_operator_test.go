package go2p_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/v-braun/go2p"
)

func TestPingPong(t *testing.T) {
	// p1log1 := CreateLogMiddleware("p1:log1")
	// p1log2 := CreateLogMiddleware("p1:log2")

	// p2log1 := CreateLogMiddleware("p1:log1")
	// p2log2 := CreateLogMiddleware("p1:log2")

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

		if p.Address() != "tcp:127.0.0.1:3378" {
			t.Fatal("unexpected address", p.Address())
			return
		}

		clientsWg.Wait()
		conn1.Send(go2p.NewMessageFromData([]byte("hello")), p.Address())
	})

	conn2.OnPeer(func(p *go2p.Peer) {
		clientsWg.Done()
	})

	conn1.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		assert.Equal(t, "hello back", m.PayloadGetString())
		fmt.Printf("from %s: %s\n", p.Address(), m.PayloadGetString())
		msgWg.Done()
	})

	conn2.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		assert.Equal(t, "hello", m.PayloadGetString())
		fmt.Printf("from %s: %s\n", p.Address(), m.PayloadGetString())
		go conn2.Send(go2p.NewMessageFromData([]byte("hello back")), p.Address())
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
}
