package go2p_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

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
	net      *go2p.NetworkConnection
	addr     string
	fullAddr string
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

	return &networkConnWithAddress{net: conn1, addr: addr1, fullAddr: "tcp:" + addr1}, &networkConnWithAddress{net: conn2, addr: addr2, fullAddr: "tcp:" + addr2}
}

func startNetworks(t *testing.T, networks ...*go2p.NetworkConnection) bool {
	for _, n := range networks {
		err := n.Start()
		if !assert.NoError(t, err) {
			return false
		}
	}

	return true
}

func registerPeerErrorHandlers(t *testing.T, networks ...*go2p.NetworkConnection) {
	for i, n := range networks {
		n.OnPeerError(func(p *go2p.Peer, err error) {
			msg := fmt.Sprintf("conn%d err: %+v", i, errors.Wrap(err, "unexpected peer error"))
			fmt.Println(msg)
			t.Fatal(msg)
		})

		// n.OnPeerDisconnect(func(p *go2p.Peer) {
		// 	msg := fmt.Sprintf("conn%d disconnect: %+v\n", i, p.RemoteAddress())
		// 	fmt.Println(msg)
		// 	t.Fatal(msg)
		// })
	}
}

func TestChat(t *testing.T) {
	messages := getChatProtocoll()
	conn1, conn2 := createTestNetworks(t, go2p.EmptyRoutesTable)

	testDone := new(sync.WaitGroup)
	testDone.Add(1)

	conn1.net.OnPeer(func(p *go2p.Peer) {
		fmt.Printf("%s got peer %s\n", p.LocalAddress(), p.RemoteAddress())
		conn1.net.Send(go2p.NewMessageFromString(messages[0].in), p.RemoteAddress())
	})

	conn1.net.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		txt := m.PayloadGetString()
		assert.Equal(t, messages[0].out, txt)

		fmt.Printf("%s got %s\n", "conn1", messages[0])
		messages = messages[1:]
		if len(messages) == 0 {
			testDone.Done()
		} else {
			conn1.net.Send(go2p.NewMessageFromString(messages[0].in), p.RemoteAddress())
		}

	})

	conn2.net.OnMessage(func(p *go2p.Peer, m *go2p.Message) {
		txt := m.PayloadGetString()
		assert.Equal(t, messages[0].in, txt)

		fmt.Printf("%s got %s\n", "conn2", messages[0])
		conn2.net.Send(go2p.NewMessageFromString(messages[0].out), p.RemoteAddress())
	})

	registerPeerErrorHandlers(t, conn1.net, conn2.net)
	if !startNetworks(t, conn1.net, conn2.net) {
		return
	}

	conn1.net.ConnectTo("tcp", conn2.addr)

	testDone.Wait()

	conn1.net.Stop()
	conn2.net.Stop()
}

func TestRouting(t *testing.T) {

	wgPings := &sync.WaitGroup{}
	wgPongs := &sync.WaitGroup{}

	sendPings := 3
	sendPongs := 3

	var conn1 *networkConnWithAddress
	var conn2 *networkConnWithAddress

	wgPings.Add(sendPings)
	wgPongs.Add(sendPongs)
	conn1, conn2 = createTestNetworks(t, &map[string]func(peer *go2p.Peer){
		"ping": func(peer *go2p.Peer) {
			fmt.Printf("ping %s -> %s \n", peer.RemoteAddress(), peer.LocalAddress())
			// if peer.Address() != conn2.fullAddr {
			// 	assert.FailNow(t, "unexpected pong from addr: %s. conn1: %s conn2: %s", peer.Address(), conn1.addr, conn2.addr)
			// }

			wgPings.Done()
			if sendPongs > 0 {
				sendPongs -= 1
				conn2.net.Send(go2p.NewMessageRoutedFromString("pong", "pong"), peer.RemoteAddress())
			}
		},
		"pong": func(peer *go2p.Peer) {
			fmt.Printf("pong %s -> %s \n", peer.RemoteAddress(), peer.LocalAddress())
			// if peer.Address() != conn1.fullAddr {
			// 	assert.FailNow(t, "unexpected pong from addr: %s. conn1: %s conn2: %s", peer.Address(), conn1.addr, conn2.addr)
			// }

			wgPongs.Done()
			if sendPings > 0 {
				sendPings -= 1
				conn1.net.Send(go2p.NewMessageRoutedFromString("ping", "ping"), peer.RemoteAddress())
			}
		},
	})

	peerConnectedWg := sync.WaitGroup{}
	peerConnectedWg.Add(2)
	conn1.net.OnPeer(func(peer *go2p.Peer) {
		peerConnectedWg.Done()
		fmt.Printf("new peer on: %s with addr: %s\n", peer.LocalAddress(), peer.RemoteAddress())
	})

	conn2.net.OnPeer(func(peer *go2p.Peer) {
		peerConnectedWg.Done()
		fmt.Printf("new peer on: %s with addr: %s\n", peer.LocalAddress(), peer.RemoteAddress())
	})

	registerPeerErrorHandlers(t, conn1.net, conn2.net)
	if !startNetworks(t, conn1.net, conn2.net) {
		return
	}

	fmt.Printf("connect to %s\n", conn2.addr)
	conn1.net.ConnectTo("tcp", conn2.addr)
	peerConnectedWg.Wait()
	fmt.Printf("send to %s\n", conn2.addr)
	conn1.net.Send(go2p.NewMessageRoutedFromString("ping", "ping"), conn2.fullAddr)

	wgPings.Wait()
	wgPongs.Wait()

	conn1.net.Stop()
	conn2.net.Stop()
}

func TestLargeMessages(t *testing.T) {

	var conn1 *networkConnWithAddress
	var conn2 *networkConnWithAddress
	largeMsg := genLargeMessage(256)
	testDone := sync.WaitGroup{}
	testDone.Add(1)
	conn1, conn2 = createTestNetworks(t, &map[string]func(peer *go2p.Peer){
		"say": func(peer *go2p.Peer) {
			assert.Equal(t, largeMsg, "")
			testDone.Done()
		},
	})

	peerConnectedWg := sync.WaitGroup{}
	peerConnectedWg.Add(2)
	conn1.net.OnPeer(func(peer *go2p.Peer) {
		peerConnectedWg.Done()
	})

	conn2.net.OnPeer(func(peer *go2p.Peer) {
		peerConnectedWg.Done()
	})

	registerPeerErrorHandlers(t, conn1.net, conn2.net)
	if !startNetworks(t, conn1.net, conn2.net) {
		return
	}

	conn1.net.ConnectTo("tcp", conn2.addr)
	peerConnectedWg.Wait()
	conn1.net.Send(go2p.NewMessageRoutedFromString("say", largeMsg), conn2.fullAddr)

	testDone.Wait()

	conn1.net.Stop()
	conn2.net.Stop()
}

func genLargeMessage(chars int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, chars)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}
	return string(b)
}
