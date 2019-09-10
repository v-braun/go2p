package go2p

import (
	"sync"
)

// DefaultPeerStore is a basic implementation of a PeerStore
// It limits simultaneously connected peers to a configured capacity
type peers struct {
	peers []*Peer
	mutex *sync.Mutex
}

// NewDefaultPeerStore creates a new basic PeerStore that limits
// simultaneously connected peers by the provided capacity
func newPeers() *peers {
	p := new(peers)
	p.peers = make([]*Peer, 0)
	p.mutex = new(sync.Mutex)

	return p
}

// AddPeer adds the given peer to the store
func (p *peers) add(peer *Peer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.peers = append(p.peers, peer)
}

// RemovePeer will remove the given peer from the store
func (p *peers) rm(peer *Peer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	peerIdx := -1
	for i, p := range p.peers {
		if p == peer {
			peerIdx = i
			break
		}
	}

	if peerIdx != -1 {
		p.peers = append(p.peers[:peerIdx], p.peers[peerIdx+1:]...)
	}
}

// IteratePeer will call the given handler for each peer
func (p *peers) iteratePeer(handler func(peer *Peer)) {
	p.mutex.Lock()
	peersCopy := make([]*Peer, len(p.peers))
	copy(peersCopy, p.peers)
	p.mutex.Unlock()

	for _, p := range peersCopy {
		handler(p)
	}
}

func (p *peers) lock(addr string, handler func(peer *Peer)) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, p := range p.peers {
		if p.io.adapter.RemoteAddress() == addr {
			handler(p)
			return
		}
	}
}
