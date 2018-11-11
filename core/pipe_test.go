package core

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"github.com/v-braun/go2p/core/mocks"
)

func createMiddleware(idx int) *Middleware {
	f := func(pipe Pipe, msg Message) (MiddlewareResult, error) {
		s := msg.PayloadGetString()
		s += " " + strconv.Itoa(idx)
		s = strings.TrimSpace(s)
		msg.PayloadSetString(s)
		return Next, nil
	}

	return NewMiddleware(fmt.Sprintf("m%d", idx), f)
}
func createMiddlewares(amount int) middlewareList {
	res := middlewareList{}
	for i := 0; i < amount; i++ {
		m := createMiddleware(i)
		res = append(res, m)
	}

	res.arrange()
	return res
}

func TestProcessPipeSend(t *testing.T) {
	actions := createMiddlewares(5)

	p := newPipe(nil, actions, Send, 0)

	msg := &message{}

	err := p.process(msg)
	assert.NoError(t, err)

	content := msg.PayloadGetString()
	assert.Equal(t, "0 1 2 3 4", content)

}
func TestProcessPipeReciev(t *testing.T) {
	actions := createMiddlewares(5)

	p := newPipe(nil, actions, Receive, 0)

	msg := &message{}

	err := p.process(msg)
	assert.NoError(t, err)

	content := msg.PayloadGetString()
	assert.Equal(t, "4 3 2 1 0", content)

}

func TestPipeRecievSend(t *testing.T) {
	actions := createMiddlewares(5)

	actions[2].Execute = func(pipe Pipe, msg Message) (MiddlewareResult, error) {

		content := msg.PayloadGetString()
		assert.Equal(t, "4 3", content)

		msgAnswer := &message{}
		msgAnswer.PayloadSetString("answer")
		pipe.Send(msgAnswer)
		contentAsw := msgAnswer.PayloadGetString()

		assert.Equal(t, "answer 3 4", contentAsw)

		msg.PayloadSetString(content + " " + contentAsw)

		return Next, nil
	}

	adapter := new(mocks.Adapter)
	adapter.On("Write", mock.Anything).Return(nil)
	peer := newPeer(adapter)
	p := newPipe(peer, actions, Receive, 0)

	msg := &message{}

	err := p.process(msg)
	assert.NoError(t, err)

	content := msg.PayloadGetString()
	assert.Equal(t, "4 3 answer 3 4 1 0", content)

}

func TestPipeSendSendReceive(t *testing.T) {
	actions := createMiddlewares(5)

	actions[2].Execute = func(pipe Pipe, msg Message) (MiddlewareResult, error) {

		content := msg.PayloadGetString()
		assert.Equal(t, "0 1", content)

		// intercept current rq and send another one
		interceptMsg := &message{}
		interceptMsg.PayloadSetString("intercept")
		pipe.Send(interceptMsg)
		interceptContent := interceptMsg.PayloadGetString()
		assert.Equal(t, "intercept 3 4", interceptContent)

		// wait for a response
		interceptResMsg, _ := pipe.Receive()
		interceptResContent := interceptResMsg.PayloadGetString()
		assert.Equal(t, "MSG 4 3", interceptResContent)

		msg.PayloadSetString(content + " [" + interceptResContent + "]")

		return Next, nil
	}

	adapter := new(mocks.Adapter)
	adapter.On("Write", mock.Anything).Return(nil)
	readCall := adapter.On("Read", mock.Anything)
	adapter.SetupReadWithResponse(readCall, "MSG")
	adapter.SetupClose(readCall, DisconnectedError)

	peer := newPeer(adapter)
	peer.run()
	p := newPipe(peer, actions, Send, 0)

	msg := &message{}

	err := p.process(msg)
	assert.NoError(t, err)

	content := msg.PayloadGetString()
	assert.Equal(t, "0 1 [MSG 4 3] 3 4", content)

	peer.shutdown()

}
