package go2p

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createFakedPeer(middlewares middlewares) *Peer {
	adapter := new(MockAdapter)
	adapter.On("Address").Return("mock")
	p := newPeer(adapter, middlewares)
	return p
}

func createMiddleware(idx int) *Middleware {
	f := func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
		s := msg.PayloadGetString()
		s += " " + strconv.Itoa(idx)
		s = strings.TrimSpace(s)
		msg.PayloadSetString(s)
		return Next, nil
	}

	return NewMiddleware(fmt.Sprintf("m%d", idx), f)
}
func createMiddlewares(amount int) middlewares {
	res := middlewares{}
	for i := 0; i < amount; i++ {
		m := createMiddleware(i)
		res = append(res, m)
	}

	result := newMiddlewares(res...)

	return result
}

func TestProcessPipeSend(t *testing.T) {
	actions := createMiddlewares(5)

	p := newPipe(createFakedPeer(actions), actions, Send, 0)

	msg := NewMessage()

	err := p.process(msg)
	assert.NoError(t, err)

	content := msg.PayloadGetString()
	assert.Equal(t, "0 1 2 3 4", content)

}
func TestProcessPipeReciev(t *testing.T) {
	actions := createMiddlewares(5)

	p := newPipe(createFakedPeer(actions), actions, Receive, 0)

	msg := NewMessage()

	err := p.process(msg)
	assert.NoError(t, err)

	content := msg.PayloadGetString()
	assert.Equal(t, "4 3 2 1 0", content)

}

func TestPipeRecievSend(t *testing.T) {
	actions := createMiddlewares(5)

	actions[2].Execute = func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {

		content := msg.PayloadGetString()
		assert.Equal(t, "4 3", content)

		msgAnswer := NewMessage()
		msgAnswer.PayloadSetString("answer")
		pipe.Send(msgAnswer)
		contentAsw := msgAnswer.PayloadGetString()

		assert.Equal(t, "answer 3 4", contentAsw)

		msg.PayloadSetString(content + " " + contentAsw)

		return Next, nil
	}

	adapter := new(MockAdapter)
	adapter.On("Address").Return("mock")
	readCall := adapter.On("ReadMessage")
	adapter.SetupReadWithResponse(readCall, "")
	adapter.SetupClose(readCall, DisconnectedError)

	adapter.On("WriteMessage", mock.Anything).Return(nil)
	peer := newPeer(adapter, actions)
	peer.start()
	p := newPipe(peer, actions, Receive, 0)

	msg := NewMessage()

	err := p.process(msg)
	assert.NoError(t, err)

	content := msg.PayloadGetString()
	assert.Equal(t, "4 3 answer 3 4 1 0", content)

	peer.stop()

}

// func TestPipeSendSendReceive(t *testing.T) {
// 	done := &sync.WaitGroup{}
// 	done.Add(2)

// 	actions := createMiddlewares(5)

// 	actions[2].Execute = func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {

// 		content := msg.PayloadGetString()
// 		assert.Equal(t, "0 1", content)

// 		// intercept current rq and send another one
// 		interceptMsg := NewMessage()
// 		interceptMsg.PayloadSetString("intercept")
// 		pipe.Send(interceptMsg)
// 		interceptContent := interceptMsg.PayloadGetString()
// 		assert.Equal(t, "intercept 3 4", interceptContent)

// 		// wait for a response
// 		interceptResMsg, _ := pipe.Receive()
// 		interceptResContent := interceptResMsg.PayloadGetString()
// 		assert.Equal(t, "MSG 4 3", interceptResContent)

// 		msg.PayloadSetString(content + " [" + interceptResContent + "]")

// 		done.Done()

// 		return Next, nil
// 	}

// 	adapter := new(MockAdapter)
// 	adapter.On("WriteMessage", mock.Anything).Return(nil)
// 	readCall := adapter.On("ReadMessage")

// 	adapter.SetupReadWithResponse(readCall, "MSG")
// 	adapter.SetupClose(readCall, DisconnectedError)

// 	peer := newPeer(adapter, actions)
// 	peer.start()
// 	p := newPipe(peer, actions, Send, 0)

// 	msg := NewMessage()

// 	err := p.process(msg)
// 	assert.NoError(t, err)

// 	content := msg.PayloadGetString()
// 	assert.Equal(t, "0 1 [MSG 4 3] 3 4", content)

// 	done.Wait()
// 	peer.stop()

// }
