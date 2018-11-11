package core

type errorConstant string

func (e errorConstant) Error() string { return string(e) }

// DisconnectedError represents Error when a peer is disconnected
const DisconnectedError = errorConstant("disconnected")

type Adapter interface {
	LocalAddr() string
	RemoteAddr() string

	Write(data []byte) error
	Read(buffer []byte) (int, error)

	Close()
}

type AdapterProvider interface {
	OnConnection() <-chan Adapter
	OnErr() <-chan error
	Close()
}

type Dialer interface {
	AdapterProvider
	Dial(addr string) <-chan struct{}
}

type Listener interface {
	AdapterProvider
	Listen(addr string)
}

type PeerStore interface {
	List() []Peer
	OnNewPeer(cb func(p Peer))
	OnLostPeer(cb func(p Peer))

	Connect(remoteAddr string)
	Send(p Peer, msg Message) <-chan struct{}
	Boradcast(msg Message) <-chan struct{}

	// test
	handlePeerError(p Peer, err error)
}

type PeerMonitor interface {
	ShouldAdd(p Peer) bool
	ShouldRemove(p Peer) bool
	Close()
}

type NetworkInterface interface {
	Start()
	Stop()

	OnError(handler func(err error))
	Peers() PeerStore
}

type PipeOperation int

const (
	Send    PipeOperation = iota
	Receive PipeOperation = iota
)
