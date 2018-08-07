package p2p

import (
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"
)

type PeerManager struct {
	neighbors *sync.Map // map[peer.ID]*Peer

	subs map[MessageType]map[string]chan IncomingMessage
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
	c := make(chan IncomingMessage, 1024)
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
