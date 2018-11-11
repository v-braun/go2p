package mocks

import (
	"encoding/binary"

	"github.com/stretchr/testify/mock"
)

func (a *Adapter) SetupReadWithResponse(c *mock.Call, response string) chan struct{} {
	result := make(chan struct{})
	timeCalled := 0
	payload := []byte(response)

	c.Run(func(arg mock.Arguments) {
		buffer := arg.Get(0).([]byte)
		if timeCalled == 0 {
			timeCalled++
			binary.BigEndian.PutUint32(buffer, uint32(len(payload)))
			c.Return(len(buffer), nil)
		} else if timeCalled == 1 {
			timeCalled++
			for i := 0; i < len(payload); i++ {
				buffer[i] = payload[i]
			}

			c.Return(len(buffer), nil)
			close(result)
		} else {
			c.Return(0, nil)
		}
	})

	return result
}

func (a *Adapter) SetupReadWithNOP(c *mock.Call) {
	c.Run(func(arg mock.Arguments) {
		c.Return(0, nil)
	})
}

func (a *Adapter) SetupClose(readCall *mock.Call, res error) {
	a.On("Close").Return().Run(func(arg mock.Arguments) {
		readCall.Run(func(arg mock.Arguments) {
			readCall.Return(0, res)
		})
	})
}
