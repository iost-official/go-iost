package p2p

import (
	"sync"
	"time"

	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/uber-go/atomic"
)

var (
	clearInactivePeerInterval = 5 * time.Second
)

const (
	maxNeighborCount = 32 // TODO: configurable

	incomingMsgChanSize = 1024
)

type PeerManager struct {
	neighbors *sync.Map // map[peer.ID]*Peer
	peerCount atomic.Uint64
	subs      map[MessageType]map[string]chan IncomingMessage
	quitCh    chan struct{}
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		neighbors: new(sync.Map),
		subs:      make(map[MessageType]map[string]chan IncomingMessage),
		quitCh:    make(chan struct{}),
	}
}

func (pm *PeerManager) Start() {
	go pm.clearInactivePeer()
}

func (pm *PeerManager) Stop() {
	close(pm.quitCh)
}

func (pm *PeerManager) clearInactivePeer() {
	ticker := time.NewTicker(clearInactivePeerInterval)
	for {
		select {
		case <-pm.quitCh:
			// log
			return
		case <-ticker.C:
			pm.neighbors.Range(func(k, v interface{}) bool {
				peer, ok := v.(*Peer)
				if !ok || peer.Inactive() {
					pm.neighbors.Delete(k)
					pm.peerCount.Dec()
				}
				return true
			})
		}
	}
}

func (pm *PeerManager) AddPeer(s libnet.Stream) {
	if pm.peerCount.Load() >= maxNeighborCount {
		s.Close()
		return
	}
	remotePID := s.Conn().RemotePeer()
	if p, exist := pm.neighbors.Load(remotePID); exist {
		old, ok := p.(*Peer)
		if ok && !old.Inactive() {
			s.Close()
			return
		}
		pm.neighbors.Delete(remotePID)
		pm.peerCount.Dec()
	}
	peer := NewPeer(s)
	peer.Start()
	pm.neighbors.Store(remotePID, peer)
	pm.peerCount.Inc()
}

func (pm *PeerManager) RemovePeer(peerID peer.ID) {
	if p, exist := pm.neighbors.Load(peerID); exist {
		pm.neighbors.Delete(peerID)
		pm.peerCount.Dec()
		if peer, ok := p.(*Peer); ok && !peer.Inactive() {
			peer.Stop()
		}
	}
}

func (pm *PeerManager) Broadcast(data []byte, typ MessageType, mp MessagePriority) {
	msg := newP2PMessage(100, typ, 1, 0, data)
	pm.neighbors.Range(func(k, v interface{}) bool {
		peer, ok := v.(*Peer)
		if !ok {
			return true
		}
		peer.SendMessage(msg, mp)
		return true
	})
}

func (pm *PeerManager) SendToPeer(peerID peer.ID, data []byte, typ MessageType, mp MessagePriority) {
	msg := newP2PMessage(100, typ, 1, 0, data)
	v, ok := pm.neighbors.Load(peerID)
	if !ok {
		// log
		return
	}
	peer, ok := v.(*Peer)
	if !ok {
		// log
		return
	}
	peer.SendMessage(msg, mp)
}

func (pm *PeerManager) Register(id string, mTyps ...MessageType) chan IncomingMessage {
	if len(mTyps) == 0 {
		return nil
	}
	c := make(chan IncomingMessage, incomingMsgChanSize)
	for _, typ := range mTyps {
		pm.subs[typ][id] = c
	}
	return c
}

func (pm *PeerManager) Deregister(id string, mTyps ...MessageType) {
	for _, typ := range mTyps {
		delete(pm.subs[typ], id)
	}
}
