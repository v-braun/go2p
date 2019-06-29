package core

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
