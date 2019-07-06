package go2p_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/v-braun/go2p"
)

type chatProtocoll []struct {
	out string
	in  string
}

type networkConnWithAddress struct {
	net  *go2p.NetworkConnection
	addr string
}

func getChatProtocoll() chatProtocoll {
	messages := []struct {
		out string
		in  string
	}{{
		out: "hello",
		in:  "hi",
	}, {
		out: "how are you",
		in:  "fine",
	}, {
		out: "nice to meet you",
		in:  "you 2",
	}, {
		out: "bye",
		in:  "see ya",
	}}

	return messages
}

func createTestNetworks(t *testing.T, routing go2p.RoutingTable) (*networkConnWithAddress, *networkConnWithAddress) {
	p1, err := freeport.GetFreePort()
	assert.NoError(t, err)

	p2, err := freeport.GetFreePort()
	assert.NoError(t, err)

	addr1 := fmt.Sprintf("127.0.0.1:%d", p1)
	addr2 := fmt.Sprintf("127.0.0.1:%d", p2)

	conn1 := go2p.NewNetworkConnectionTCP(addr1, routing)
	conn2 := go2p.NewNetworkConnectionTCP(addr2, routing)

	return &networkConnWithAddress{net: conn1, addr: addr1}, &networkConnWithAddress{net: conn2, addr: addr2}
}

func TestChat(t *testing.T) {
	messages := getChatProtocoll()
	conn1, conn2 := createTestNetworks(t, go2p.EmptyRoutesTable)

	testDone := new(sync.WaitGroup)
	testDone.Add(1)

	conn1.net.OnPeer(func(p *go2p.Peer) {
		fmt.Printf("%s got peer %s\n", "conn1", p.Address())
		conn1.net.Send(go2p.NewMessageFromString(messages[0].in), p.Address())
	})

	conn1.net.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		txt := m.PayloadGetString()
		assert.Equal(t, messages[0].out, txt)

		fmt.Printf("%s got %s\n", "conn1", messages[0])
		messages = messages[1:]
		if len(messages) == 0 {
			testDone.Done()
		} else {
			conn1.net.Send(go2p.NewMessageFromString(messages[0].in), p.Address())
		}

	})

	conn2.net.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		txt := m.PayloadGetString()
		assert.Equal(t, messages[0].in, txt)

		fmt.Printf("%s got %s\n", "conn2", messages[0])
		conn2.net.Send(go2p.NewMessageFromString(messages[0].out), p.Address())
	})

	conn1.net.OnPeerError(func(p *go2p.Peer, err error) {
		fmt.Printf("conn1 err: %+v", errors.Wrap(err, "unexpected peer error"))
	})
	conn2.net.OnPeerError(func(p *go2p.Peer, err error) {
		fmt.Printf("conn2 err: %+v", errors.Wrap(err, "unexpected peer error"))
	})

	err := conn1.net.Start()
	if !assert.NoError(t, err) {
		return
	}

	err = conn2.net.Start()
	if !assert.NoError(t, err) {
		return
	}

	conn1.net.ConnectTo("tcp", conn2.addr)

	fmt.Printf("start waiting ...\n")
	testDone.Wait()

	fmt.Printf("start closing conn1 ...\n")
	conn1.net.Stop()
	fmt.Printf("start closing conn2 ...\n")
	conn2.net.Stop()
}

// func TestRouting(t *testing.T) {

// 	wgPings := &sync.WaitGroup{}
// 	wgPongs := &sync.WaitGroup{}

// 	sendPings := 3
// 	sendPongs := 3

// 	var conn1 *networkConnWithAddress
// 	var conn2 *networkConnWithAddress

// 	wgPings.Add(sendPings)
// 	wgPongs.Add(sendPongs)
// 	conn1, conn2 = createTestNetworks(t, &map[string]func(peer *go2p.Peer){
// 		"ping": func(peer *go2p.Peer) {
// 			if peer.Address() != conn2.addr {
// 				assert.FailNow(t, "unexpected pong from addr: %s. conn1: %s conn2: %s", peer.Address(), conn1.addr, conn2.addr)
// 			}

// 			wgPings.Done()
// 			if sendPongs > 0 {
// 				sendPongs -= 1
// 				conn2.net.Send(go2p.NewMessageRoutedFromString("pong"), "pong")
// 			}
// 		},
// 		"pong": func(peer *go2p.Peer) {
// 			if peer.Address() != conn1.addr {
// 				assert.FailNow(t, "unexpected pong from addr: %s. conn1: %s conn2: %s", peer.Address(), conn1.addr, conn2.addr)
// 			}

// 			wgPongs.Done()
// 			if sendPings > 0 {
// 				sendPings -= 1
// 				conn2.net.Send(go2p.NewMessageRoutedFromString("ping"), "ping")
// 			}
// 		},
// 	})

// }
