package go2p_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/v-braun/go2p"
)

func TestChat(t *testing.T) {

	messages := make([]struct {
		out string
		in  string
	}, 4)

	messages[0].in = "hello"
	messages[0].out = "hi"
	messages[1].in = "how are you"
	messages[1].out = "fine"
	messages[2].in = "nice to meet you"
	messages[2].out = "you 2"
	messages[3].in = "bye"
	messages[3].out = "see ya"

	p1, err := freeport.GetFreePort()
	assert.NoError(t, err)

	p2, err := freeport.GetFreePort()
	assert.NoError(t, err)

	addr1 := fmt.Sprintf("127.0.0.1:%d", p1)
	addr2 := fmt.Sprintf("127.0.0.1:%d", p2)

	conn1 := go2p.NewNetworkConnectionTCP(addr1)
	conn2 := go2p.NewNetworkConnectionTCP(addr2)

	testDone := new(sync.WaitGroup)
	testDone.Add(1)

	conn1.OnPeer(func(p *go2p.Peer) {
		conn1.Send(go2p.NewMessageFromString(messages[0].in), p.Address())
	})

	conn1.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		txt := m.PayloadGetString()
		assert.Equal(t, messages[0].out, txt)

		messages = messages[1:]
		if len(messages) == 0 {
			testDone.Done()
		} else {
			conn1.Send(go2p.NewMessageFromString(messages[0].in), p.Address())
		}

	})

	conn2.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		txt := m.PayloadGetString()
		assert.Equal(t, messages[0].in, txt)

		conn2.Send(go2p.NewMessageFromString(messages[0].out), p.Address())
	})

	conn1.Start()
	conn2.Start()

	conn1.ConnectTo("tcp", addr2)

	testDone.Wait()
	conn1.Stop()
	conn2.Stop()

}
