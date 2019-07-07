package go2p

import (
	"sync"
	"time"

	"github.com/v-braun/awaiter"
)

// PeerStore is a store of peers used to manage connections to the peers.
// It can be used to keep only a limit of opened connections or to filter specific
// peers
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

// DefaultPeerStore is a basic implementation of a PeerStore
// It limits simultaneously connected peers to a configured capacity
type DefaultPeerStore struct {
	peers    []*Peer
	mutex    *sync.Mutex
	emitter  *eventEmitter
	awaiter  awaiter.Awaiter
	capacity int
}

// NewDefaultPeerStore creates a new basic PeerStore that limits
// simultaneously connected peers by the provided capacity
func NewDefaultPeerStore(capacity int) PeerStore {
	ps := new(DefaultPeerStore)
	ps.peers = make([]*Peer, 0)
	ps.mutex = new(sync.Mutex)
	ps.emitter = newEventEmitter()
	ps.capacity = capacity
	ps.awaiter = awaiter.New()

	return ps
}

// AddPeer adds the given peer to the store
func (ps *DefaultPeerStore) AddPeer(peer *Peer) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.peers = append(ps.peers, peer)
	ps.emitter.EmitAsync("add-peer", peer)
	return nil
}

// OnPeerAdd registers the given handler and calls it when a new peer should be added
func (ps *DefaultPeerStore) OnPeerAdd(handler func(peer *Peer)) {
	ps.emitter.On("add-peer", func(args []interface{}) {
		handler(args[0].(*Peer))
	})
}

// OnPeerWantRemove registers the given handler and calls it when
// a peer should be removed
func (ps *DefaultPeerStore) OnPeerWantRemove(handler func(peer *Peer)) {
	ps.emitter.On("remove-peer", func(args []interface{}) {
		handler(args[0].(*Peer))
	})
}

// RemovePeer will remove the given peer from the store
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

// IteratePeer will call the given handler for each peer
func (ps *DefaultPeerStore) IteratePeer(handler func(peer *Peer)) {
	ps.mutex.Lock()
	peersCopy := make([]*Peer, len(ps.peers))
	copy(peersCopy, ps.peers)
	ps.mutex.Unlock()

	for _, p := range peersCopy {
		handler(p)
	}
}

// Start the background routines that monitors all connected peers
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

// Stop the background monitoring
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

// LockPeer is used to search for a peer, locks it and calls the handler
// after the handler was called the peer will be unlocked
func (ps *DefaultPeerStore) LockPeer(addr string, handler func(peer *Peer)) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for _, p := range ps.peers {
		if p.io.adapter.RemoteAddress() == addr {
			handler(p)
			return
		}
	}
}
