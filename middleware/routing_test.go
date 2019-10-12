package middleware_test

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/internal"
	"github.com/v-braun/go2p/middleware"
)

func TestRoute(t *testing.T) {
	sut1 := internal.CreateTCPSUT(t)
	sut2 := internal.CreateTCPSUT(t)

	expectedMsg := "bar"
	actualMsg := ""
	sut1.Wg.Add(1)
	sut1.PrependMiddleware(middleware.Route("foo", func(peer *core.Peer, msg *core.Message) {
		actualMsg = msg.PayloadGetString()
		sut1.Wg.Done()
	}))
	sut2.PrependMiddleware(middleware.Route("foo", func(peer *core.Peer, msg *core.Message) {

	}))

	sut2.Wg.Add(1)
	sut2.OnPeer(func(p *core.Peer) {
		sut2.Wg.Done()
	})

	sut1.Start()
	defer sut1.Stop()
	sut2.Start()
	defer sut2.Stop()

	sut2.ConnectTo("tcp", sut1.Addr)
	sut2.Wg.Wait()

	msg := core.NewMessageFromString(expectedMsg)
	msg = middleware.AssignRouteToMessage("foo", msg)
	sut2.Send(msg, sut1.FullAddr)

	sut1.Wg.Wait()

	assert.Equal(t, expectedMsg, actualMsg)
}
