package core

type Extension interface {
	Install(n *Network) error
	Uninstall()
}

type Dialer interface {
	Dial(network string, addr string) error
}

type Listener interface {
	OnPeer(handler func(p Conn))
}

type ErrorProvider interface {
	OnError(handler func(err error))
}

type Operator interface {
	Extension
	Dialer
	Listener
}
