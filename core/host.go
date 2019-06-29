package core

type Host interface {
}

type host struct {
	listener []Listener
	dialer   Dialer
	peers    PeerStore
}

func (s *host) startHost() {
}
