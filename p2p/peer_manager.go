package p2p

import (
	"sync"
	"time"

	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
)

var (
	clearInactivePeerInterval = 5 * time.Second
)

const (
	maxPeerCount = 32 // TODO: configurable

	incomingMsgChanSize = 1024
)

type PeerManager struct {
	peers     map[peer.ID]*Peer
	peerCount int
	peerMutex sync.RWMutex

	subs   map[MessageType]map[string]chan IncomingMessage
	quitCh chan struct{}
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers:  make(map[peer.ID]*Peer),
		subs:   make(map[MessageType]map[string]chan IncomingMessage),
		quitCh: make(chan struct{}),
	}
}

func (pm *PeerManager) Start() {
}

func (pm *PeerManager) Stop() {
	close(pm.quitCh)
}

func (pm *PeerManager) HandlerStream(s libnet.Stream) {
	remotePID := s.Conn().RemotePeer()
	peer := pm.GetPeer(remotePID)
	if peer == nil {
		if pm.PeerCount() >= maxPeerCount {
			s.Reset()
			return
		}
		pm.AddPeer(NewPeer(s, pm))
		return
	}

	err := peer.AddStream(s)
	if err != nil {
		s.Reset()
		return
	}
}

func (pm *PeerManager) AddPeer(p *Peer) {
	pm.peerMutex.Lock()
	defer pm.peerMutex.Unlock()

	pm.peers[p.id] = p
	p.Start()
}

func (pm *PeerManager) RemovePeer(peerID peer.ID) {
	pm.peerMutex.Lock()
	defer pm.peerMutex.Unlock()

	if peer, exist := pm.peers[peerID]; exist {
		peer.Stop()
		delete(pm.peers, peerID)
	}
}

func (pm *PeerManager) GetPeer(peerID peer.ID) *Peer {
	pm.peerMutex.RLock()
	defer pm.peerMutex.RUnlock()

	return pm.peers[peerID]
}

func (pm *PeerManager) PeerCount() int {
	pm.peerMutex.RLock()
	defer pm.peerMutex.RUnlock()

	return len(pm.peers)
}

func (pm *PeerManager) Broadcast(data []byte, typ MessageType, mp MessagePriority) {
	msg := newP2PMessage(100, typ, 1, 0, data)

	pm.peerMutex.RLock()
	defer pm.peerMutex.RUnlock()

	for _, peer := range pm.peers {
		peer.SendMessage(msg, mp)
	}
}

func (pm *PeerManager) SendToPeer(peerID peer.ID, data []byte, typ MessageType, mp MessagePriority) {
	msg := newP2PMessage(100, typ, 1, 0, data)
	peer := pm.GetPeer(peerID)
	if peer != nil {
		peer.SendMessage(msg, mp)
	}
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
