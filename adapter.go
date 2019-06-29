package go2p

type Adapter interface {
	Receive() (*Message, error)
	Send(m *Message) error
}
