package core

import (
	"sync"
)

type peers struct {
	peers map[*Peer]bool
	mutex *sync.Mutex
}

func newPeers() *peers {
	p := new(peers)
	p.peers = make(map[*Peer]bool, 0)
	p.mutex = new(sync.Mutex)

	return p
}

func (p *peers) add(peer *Peer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.peers[peer] = true
}

func (p *peers) rm(peer *Peer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.peers[peer] {
		p.peers[peer] = false
		delete(p.peers, peer)
	}
}

func (p *peers) findByAddr(addr string) *Peer {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for p := range p.peers {
		if p.RemoteAddress() == addr {
			return p
		}
	}

	return nil
}

func (p *peers) allPeers() []*Peer {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	result := make([]*Peer, 0)
	for p := range p.peers {
		result = append(result, p)
	}

	return result
}
