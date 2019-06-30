package go2p

type errorConstant string

func (e errorConstant) Error() string { return string(e) }

// DisconnectedError represents Error when a peer is disconnected
const DisconnectedError = errorConstant("disconnected")

type Adapter interface {
	ReadMessage() (*Message, error)
	WriteMessage(m *Message) error
	Close()
	Address() string
}
