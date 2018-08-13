package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/uber-go/atomic"
)

var (
	dumpRoutingTableInterval = 5 * time.Minute
	syncRoutingTableInterval = 30 * time.Second
)

const (
	maxNeighborCount  = 32 // TODO: configurable
	peerResponseCount = 20

	incomingMsgChanSize = 1024
)

type PeerManager struct {
	neighbors     map[peer.ID]*Peer
	neighborCount int
	neighborMutex sync.RWMutex

	subs   *sync.Map //  map[MessageType]map[string]chan IncomingMessage
	quitCh chan struct{}

	seeds []string

	host            host.Host
	routingTable    *kbucket.RoutingTable
	peerStore       peerstore.Peerstore
	routingFilePath string
	lastUpdateTime  atomic.Int64
}

func NewPeerManager(host host.Host) *PeerManager {
	routingTable := kbucket.NewRoutingTable(20, kbucket.ConvertPeerID(host.ID()), time.Second, host.Peerstore())
	return &PeerManager{
		neighbors:    make(map[peer.ID]*Peer),
		subs:         new(sync.Map),
		quitCh:       make(chan struct{}),
		routingTable: routingTable,
		peerStore:    host.Peerstore(),
	}
}

func (pm *PeerManager) Start() {
	pm.parseSeeds()
	pm.LoadRoutingTable()
	pm.syncRoutingTable()

	go pm.dumpRoutingTableLoop()
	go pm.syncRoutingTableLoop()

}

func (pm *PeerManager) Stop() {
	close(pm.quitCh)
}

func (pm *PeerManager) HandleStream(s libnet.Stream) {
	remotePID := s.Conn().RemotePeer()
	peer := pm.GetNeighbor(remotePID)
	if peer == nil {
		if pm.NeighborCount() >= maxNeighborCount {
			s.Reset()
			return
		}
		pm.AddNeighbor(NewPeer(s, pm))
		return
	}

	err := peer.AddStream(s)
	if err != nil {
		s.Reset()
		return
	}
}

func (pm *PeerManager) dumpRoutingTableLoop() {
	var lastSaveTime int64
	dumpRoutingTableTicker := time.NewTimer(dumpRoutingTableInterval)
	for {
		select {
		case <-pm.quitCh:
			return
		case <-dumpRoutingTableTicker.C:
			if lastSaveTime < pm.lastUpdateTime.Load() {
				pm.DumpRoutingTable()
				lastSaveTime = time.Now().Unix()
				dumpRoutingTableTicker.Reset(dumpRoutingTableInterval)
			}
		}
	}
}

func (pm *PeerManager) syncRoutingTableLoop() {
	syncRoutingTableTicker := time.NewTimer(syncRoutingTableInterval)
	for {
		select {
		case <-pm.quitCh:
			return
		case <-syncRoutingTableTicker.C:
			pm.syncRoutingTable()
			syncRoutingTableTicker.Reset(syncRoutingTableInterval)
		}
	}
}

func (pm *PeerManager) storePeer(peerID peer.ID, addr multiaddr.Multiaddr) {
	pm.peerStore.SetAddr(peerID, addr, peerstore.PermanentAddrTTL)
	pm.routingTable.Update(peerID)
	pm.lastUpdateTime.Store(time.Now().Unix())
}

func (pm *PeerManager) deletePeer(peerID peer.ID) {
	pm.peerStore.ClearAddrs(peerID)
	pm.routingTable.Remove(peerID)
	pm.lastUpdateTime.Store(time.Now().Unix())
}

func (pm *PeerManager) AddNeighbor(p *Peer) {
	p.Start()
	pm.storePeer(p.id, p.addr)

	pm.neighborMutex.Lock()
	defer pm.neighborMutex.Unlock()

	pm.neighbors[p.id] = p
}

func (pm *PeerManager) RemoveNeighbor(peerID peer.ID) {
	pm.deletePeer(peerID)

	pm.neighborMutex.Lock()
	defer pm.neighborMutex.Unlock()

	if peer, exist := pm.neighbors[peerID]; exist {
		peer.Stop()
		delete(pm.neighbors, peerID)
	}
}

func (pm *PeerManager) GetNeighbor(peerID peer.ID) *Peer {
	pm.neighborMutex.RLock()
	defer pm.neighborMutex.RUnlock()

	return pm.neighbors[peerID]
}

func (pm *PeerManager) NeighborCount() int {
	pm.neighborMutex.RLock()
	defer pm.neighborMutex.RUnlock()

	return len(pm.neighbors)
}

func (pm *PeerManager) DumpRoutingTable() {
	file, err := os.Create(pm.routingFilePath)
	if err != nil {
		// log
		return
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("# %s\n", time.Now().String()))
	for _, peerID := range pm.routingTable.ListPeers() {
		for _, addr := range pm.peerStore.Addrs(peerID) {
			line := fmt.Sprintf("%s/ipfs/%s\n", addr.String(), peerID.Pretty())
			file.WriteString(line)
		}
	}
}

func (pm *PeerManager) LoadRoutingTable() {
	file, err := os.Open(pm.routingFilePath)
	if err != nil {
		// log
		return
	}
	defer file.Close()
	br := bufio.NewReader(file)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		peerID, addr, err := parseMultiaddr(line)
		if err != nil {
			continue
		}
		pm.storePeer(peerID, addr)
	}
}

func (pm *PeerManager) syncRoutingTable() {
	pm.Broadcast(nil, RoutingTableQuery, UrgentMessage)
	neighborCount := pm.NeighborCount()
	if neighborCount >= maxNeighborCount {
		return
	}
	allPeerIDs := pm.routingTable.ListPeers()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	perm := r.Perm(len(allPeerIDs))
	for i := 0; i < len(perm) && i < maxNeighborCount-neighborCount; i++ {
		peerID := allPeerIDs[perm[i]]
		if pm.GetNeighbor(peerID) != nil {
			continue
		}
		stream, err := pm.host.NewStream(context.Background(), peerID, protocolID)
		if err != nil {
			continue
		}
		pm.HandleStream(stream)
		pm.SendToPeer(peerID, nil, RoutingTableQuery, UrgentMessage)
	}
}

func (pm *PeerManager) parseSeeds() {
	for _, seed := range pm.seeds {
		peerID, addr, err := parseMultiaddr(seed)
		if err != nil {
			continue
		}
		pm.storePeer(peerID, addr)
	}
}

func (pm *PeerManager) Broadcast(data []byte, typ MessageType, mp MessagePriority) {
	msg := newP2PMessage(100, typ, 1, 0, data)

	pm.neighborMutex.RLock()
	defer pm.neighborMutex.RUnlock()

	for _, peer := range pm.neighbors {
		peer.SendMessage(msg, mp)
	}
}

func (pm *PeerManager) SendToPeer(peerID peer.ID, data []byte, typ MessageType, mp MessagePriority) {
	msg := newP2PMessage(100, typ, 1, 0, data)
	peer := pm.GetNeighbor(peerID)
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
		m, _ := pm.subs.LoadOrStore(typ, new(sync.Map))
		m.(*sync.Map).Store(id, c)
	}
	return c
}

func (pm *PeerManager) Deregister(id string, mTyps ...MessageType) {
	for _, typ := range mTyps {
		if m, exist := pm.subs.Load(typ); exist {
			m.(*sync.Map).Delete(id)
		}
	}
}

func (pm *PeerManager) handleRoutingTableQuery(peerID peer.ID) {
	peerIDs := pm.routingTable.NearestPeers(kbucket.ConvertPeerID(peerID), peerResponseCount)
	peerInfo := make([]peerstore.PeerInfo, 0, len(peerIDs))
	for _, id := range peerIDs {
		info := pm.peerStore.PeerInfo(id)
		if len(info.Addrs) > 0 {
			peerInfo = append(peerInfo, info)
		}
	}
	bytes, err := json.Marshal(peerInfo)
	if err != nil {
		return
	}
	pm.SendToPeer(peerID, bytes, RoutingTableResponse, UrgentMessage)
}

func (pm *PeerManager) handleRoutingTableResponse(msg *p2pMessage) {
	data, err := msg.data()
	if err != nil {
		return
	}
	peerInfos := make([]peerstore.PeerInfo, 0)
	err = json.Unmarshal(data, &peerInfos)
	if err != nil {
		return
	}
	for _, peerInfo := range peerInfos {
		if len(peerInfo.Addrs) > 0 {
			pm.storePeer(peerInfo.ID, peerInfo.Addrs[0])
		}
	}
}

func (pm *PeerManager) HandleMessage(msg *p2pMessage, peerID peer.ID) {
	data, err := msg.data()
	if err != nil {
		return
	}
	switch msg.messageType() {
	case RoutingTableQuery:
		pm.handleRoutingTableQuery(peerID)
	case RoutingTableResponse:
		pm.handleRoutingTableResponse(msg)
	default:
		inMsg := NewIncomingMessage(peerID, data, msg.messageType())
		if m, exist := pm.subs.Load(msg.messageType()); exist {
			m.(*sync.Map).Range(func(k, v interface{}) bool {
				select {
				case v.(chan IncomingMessage) <- *inMsg:
				default:
					// log
				}
				return true
			})
		}
	}
}
