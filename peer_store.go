package go2p

import (
	"sync"
	"time"

	"github.com/v-braun/awaiter"
)

type PeerStore interface {
	AddPeer(peer *Peer) error
	RemovePeer(peer *Peer)
	IteratePeer(handler func(peer *Peer))
	OnPeerAdd(handler func(peer *Peer))
	OnPeerWantRemove(handler func(peer *Peer))
	LockPeer(addr string, handler func(peer *Peer))
	Start()
	Stop()
}

type DefaultPeerStore struct {
	peers    []*Peer
	mutex    *sync.Mutex
	emitter  *eventEmitter
	awaiter  awaiter.Awaiter
	capacity int
}

func NewDefaultPeerStore(capacity int) PeerStore {
	ps := new(DefaultPeerStore)
	ps.peers = make([]*Peer, 0)
	ps.mutex = new(sync.Mutex)
	ps.emitter = newEventEmitter()
	ps.capacity = capacity
	ps.awaiter = awaiter.New()

	return ps
}

func (ps *DefaultPeerStore) AddPeer(peer *Peer) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.peers = append(ps.peers, peer)
	ps.emitter.EmitAsync("add-peer", peer)
	return nil
}

func (ps *DefaultPeerStore) OnPeerAdd(handler func(peer *Peer)) {
	ps.emitter.On("add-peer", func(args []interface{}) {
		handler(args[0].(*Peer))
	})
}
func (ps *DefaultPeerStore) OnPeerWantRemove(handler func(peer *Peer)) {
	ps.emitter.On("remove-peer", func(args []interface{}) {
		handler(args[0].(*Peer))
	})
}

func (ps *DefaultPeerStore) RemovePeer(peer *Peer) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	peerIdx := -1
	for i, p := range ps.peers {
		if p == peer {
			peerIdx = i
			break
		}
	}

	if peerIdx != -1 {
		ps.peers = append(ps.peers[:peerIdx], ps.peers[peerIdx+1:]...)
	}
}
func (ps *DefaultPeerStore) IteratePeer(handler func(peer *Peer)) {
	ps.mutex.Lock()
	peersCopy := make([]*Peer, len(ps.peers))
	copy(peersCopy, ps.peers)
	ps.mutex.Unlock()

	for _, p := range peersCopy {
		handler(p)
	}
}

func (ps *DefaultPeerStore) Start() {
	ps.awaiter.Go(func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ps.awaiter.CancelRequested():
				return
			case <-ticker.C:
				ps.checkCapa()
			}
		}
	})
}

func (ps *DefaultPeerStore) Stop() {
	ps.awaiter.Cancel()
	ps.awaiter.AwaitSync()

}

func (ps *DefaultPeerStore) checkCapa() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if len(ps.peers) <= ps.capacity {
		return
	}

	peer2Rem := ps.peers[0]

	ps.emitter.EmitAsync("remove-peer", peer2Rem)
}

func (ps *DefaultPeerStore) LockPeer(addr string, handler func(peer *Peer)) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for _, p := range ps.peers {
		if p.io.adapter.Address() == addr {
			handler(p)
			return
		}
	}
}
