package go2p

import (
	"sync"
	"time"

	"github.com/olebedev/emitter"
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
	emitter  *emitter.Emitter
	stopper  *StopSignal
	capacity int
}

func NewDefaultPeerStore(capacity int) PeerStore {
	ps := new(DefaultPeerStore)
	ps.peers = make([]*Peer, 0)
	ps.mutex = new(sync.Mutex)
	ps.emitter = emitter.New(10)
	ps.emitter.Use("*", emitter.Void)
	ps.capacity = capacity
	ps.stopper = NewStopSignal()

	return ps
}

func (ps *DefaultPeerStore) AddPeer(peer *Peer) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.peers = append(ps.peers, peer)
	go ps.emitter.Emit("add-peer", peer)
	return nil
}

func (ps *DefaultPeerStore) OnPeerAdd(handler func(peer *Peer)) {
	ps.emitter.On("add-peer", func(ev *emitter.Event) {
		handler(ev.Args[0].(*Peer))
	})
}
func (ps *DefaultPeerStore) OnPeerWantRemove(handler func(peer *Peer)) {
	ps.emitter.On("remove-peer", func(ev *emitter.Event) {
		handler(ev.Args[0].(*Peer))
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

	ps.stopper.Add()
	go func(ps *DefaultPeerStore) {
		defer ps.stopper.Done()

		ticker := time.NewTicker(time.Second * 10)

		for {
			select {
			case <-ps.stopper.Stopped():
				return
			case <-ticker.C:
				ps.checkCapa()
			}
		}

	}(ps)
}

func (ps *DefaultPeerStore) Stop() {
	ps.stopper.Stop()
	ps.stopper.Wait()

}

func (ps *DefaultPeerStore) checkCapa() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if len(ps.peers) <= ps.capacity {
		return
	}

	peer2Rem := ps.peers[0]

	go ps.emitter.Emit("remove-peer", peer2Rem)
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
