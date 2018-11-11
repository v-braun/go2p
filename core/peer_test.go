package core

import (
	"encoding/binary"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/v-braun/go2p/core/mocks"
)

func TestReceiveFromAdapter(t *testing.T) {
	adapter := new(mocks.Adapter)

	readCall := adapter.On("Read", mock.Anything)
	adapter.SetupReadWithResponse(readCall, "hello")
	adapter.SetupClose(readCall, DisconnectedError)

	adapter.On("Close").Return().Run(func(arg mock.Arguments) {
		readCall.RunFn = nil
		readCall.Return(0, DisconnectedError)
	})

	sut := newPeer(adapter)
	sut.run()

	msg := <-sut.receiveMsg
	assert.Equal(t, "hello", msg.PayloadGetString())

	sut.shutdown()

	<-sut.stop
}

func TestSendToAdapter(t *testing.T) {
	adapter := new(mocks.Adapter)

	readCall := adapter.On("Read", mock.Anything)
	adapter.SetupReadWithNOP(readCall)
	adapter.SetupClose(readCall, DisconnectedError)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	var writtenData []byte
	adapter.On("Write", mock.Anything).Return(nil).Run(func(arg mock.Arguments) {
		writtenData = arg.Get(0).([]byte)
		wg.Done()
	})

	sut := newPeer(adapter)
	sut.run()

	msg := new(message)
	msg.PayloadSetString("hello")

	sut.send(msg)

	wg.Wait()

	size := int(binary.BigEndian.Uint32(writtenData[:4]))
	content := string(writtenData[4:])

	assert.Equal(t, "hello", content)
	assert.Equal(t, len("hello"), size)

	sut.shutdown()

	<-sut.stop
}

func TestReceiveError(t *testing.T) {
	adapter := new(mocks.Adapter)

	readCall := adapter.On("Read", mock.Anything)
	readCall.Return(0, errors.New("failed"))

	adapter.On("Close").Return().Run(func(arg mock.Arguments) {
		readCall.RunFn = nil
		readCall.Return(0, DisconnectedError)
	})

	sut := newPeer(adapter)
	sut.run()

	err := <-sut.receiveErr
	assert.Error(t, err)

	sut.shutdown()
}

func TestReceiveClose(t *testing.T) {
	adapter := new(mocks.Adapter)

	readCall := adapter.On("Read", mock.Anything)
	readCall.Return(0, DisconnectedError)

	adapter.On("Close").Return().Run(func(arg mock.Arguments) {
		readCall.RunFn = nil
		readCall.Return(0, DisconnectedError)
	})

	sut := newPeer(adapter)
	sut.run()

	sut.receive()

	sut.shutdown()

	<-sut.stop
}

func TestSendError(t *testing.T) {
	adapter := new(mocks.Adapter)

	readCall := adapter.On("Read", mock.Anything)
	adapter.SetupReadWithNOP(readCall)
	adapter.SetupClose(readCall, DisconnectedError)

	adapter.On("Write", mock.Anything).Return(errors.New("test-err"))

	sut := newPeer(adapter)
	sut.run()

	msg := new(message)
	msg.PayloadSetString("hello")

	err := sut.send(msg)
	assert.Error(t, err)

	sut.shutdown()
	<-sut.stop
}

func TestSendConnEnd(t *testing.T) {
	adapter := new(mocks.Adapter)

	readCall := adapter.On("Read", mock.Anything)
	adapter.SetupReadWithNOP(readCall)
	adapter.SetupClose(readCall, DisconnectedError)

	adapter.On("Write", mock.Anything).Return(DisconnectedError)

	sut := newPeer(adapter)
	sut.run()

	msg := new(message)
	msg.PayloadSetString("hello")

	err := sut.send(msg)
	assert.Equal(t, DisconnectedError, err)

	sut.shutdown()
	<-sut.stop
}

func TestShutdownOnReceive(t *testing.T) {
	adapter := new(mocks.Adapter)

	readCall := adapter.On("Read", mock.Anything)
	dataReaded := adapter.SetupReadWithResponse(readCall, "hello")
	adapter.SetupClose(readCall, DisconnectedError)

	sut := newPeer(adapter)
	sut.run()

	<-dataReaded

	sut.shutdown()

	<-sut.stop
}

func TestConnected(t *testing.T) {
	adapter := new(mocks.Adapter)

	readCall := adapter.On("Read", mock.Anything)
	adapter.SetupReadWithResponse(readCall, "hello")
	adapter.SetupClose(readCall, DisconnectedError)

	sut := newPeer(adapter)
	sut.run()

	connected := sut.Connected()
	assert.True(t, connected)

	sut.shutdown()

	<-sut.stop

	connected = sut.Connected()
	assert.False(t, connected)

}
