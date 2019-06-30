package go2p

import (
	"github.com/stretchr/testify/mock"
)

func (a *MockAdapter) SetupReadWithResponse(c *mock.Call, response string) chan struct{} {
	result := make(chan struct{})
	// timeCalled := 0

	c.Run(func(arg mock.Arguments) {
		// if timeCalled == 0 {
		// timeCalled++
		m := NewMessageFromData([]byte(response))
		c.Return(m, nil)
		// } else if timeCalled == 1 {
		// 	timeCalled++
		// 	m := NewMessageFromData([]byte(response))
		// 	c.Return(m, nil)
		// 	close(result)
		// } else {
		// 	c.Return(nil, nil)
		// }
	})

	return result
}

func (a *MockAdapter) SetupReadWithNOP(c *mock.Call) {
	c.Run(func(arg mock.Arguments) {
		c.Return(0, nil)
	})
}

func (a *MockAdapter) SetupClose(readCall *mock.Call, res error) {
	a.On("Close").Return().Run(func(arg mock.Arguments) {
		readCall.Run(func(arg mock.Arguments) {
			readCall.Return(0, res)
		})
	})
}
