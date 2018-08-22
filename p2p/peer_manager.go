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

	"github.com/iost-official/Go-IOS-Protocol/ilog"
)

var (
	dumpRoutingTableInterval = 5 * time.Minute
	syncRoutingTableInterval = 30 * time.Second
)

const (
	maxNeighborCount  = 32 // TODO: configurable
	bucketSize        = 20
	peerResponseCount = 20

	incomingMsgChanSize = 1024
)

// PeerManager manages all peers we connect directily.
//
// PeerManager's jobs are:
//   * holding a certain amount of peers.
//   * handling messages according to its type.
//   * discovering peers and maintaining routing table.
type PeerManager struct {
	neighbors     map[peer.ID]*Peer
	neighborCount int
	neighborMutex sync.RWMutex

	subs   *sync.Map //  map[MessageType]map[string]chan IncomingMessage
	quitCh chan struct{}

	host           host.Host
	config         *Config
	routingTable   *kbucket.RoutingTable
	peerStore      peerstore.Peerstore
	lastUpdateTime atomic.Int64
}

// NewPeerManager returns a new instance of PeerManager struct.
func NewPeerManager(host host.Host, config *Config) *PeerManager {
	routingTable := kbucket.NewRoutingTable(bucketSize, kbucket.ConvertPeerID(host.ID()), time.Second, host.Peerstore())
	return &PeerManager{
		neighbors:    make(map[peer.ID]*Peer),
		subs:         new(sync.Map),
		quitCh:       make(chan struct{}),
		routingTable: routingTable,
		host:         host,
		config:       config,
		peerStore:    host.Peerstore(),
	}
}

// Start starts peer manager's job.
func (pm *PeerManager) Start() {
	pm.parseSeeds()
	pm.LoadRoutingTable()
	pm.syncRoutingTable()

	go pm.dumpRoutingTableLoop()
	go pm.syncRoutingTableLoop()

}

// Stop stops peer manager's loop.
func (pm *PeerManager) Stop() {
	close(pm.quitCh)
}

// HandleStream handles the incoming stream.
//
// It checks whether the remote peer already exists.
// If the peer is new and the neighbor count doesn't reach the threshold, it adds the peer into the neighbor list.
// If peer already exits, just add the stream to the peer.
// In other cases, reset the stream.
func (pm *PeerManager) HandleStream(s libnet.Stream) {
	remotePID := s.Conn().RemotePeer()
	ilog.Info("new stream is coming. pid=%s, addr=%v", remotePID.Pretty(), s.Conn().RemoteMultiaddr())

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
		ilog.Info("add stream failed. err=%v, pid=%s", err, remotePID.Pretty())
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
			}
			dumpRoutingTableTicker.Reset(dumpRoutingTableInterval)
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
			ilog.Info("start sync routing table.")
			pm.syncRoutingTable()
			syncRoutingTableTicker.Reset(syncRoutingTableInterval)
		}
	}
}

// storePeer stores peer information in peerStore and routingTable. It doesn't need lock since the
// peerStore.SetAddr and routingTable.Update function are thread safe.
func (pm *PeerManager) storePeer(peerID peer.ID, addr multiaddr.Multiaddr) {
	pm.peerStore.SetAddr(peerID, addr, peerstore.PermanentAddrTTL)
	pm.routingTable.Update(peerID)
	pm.lastUpdateTime.Store(time.Now().Unix())
}

// deletePeer deletes peer information in peerStore and routingTable. It doesn't need lock since the
// peerStore.SetAddr and routingTable.Update function are thread safe.
func (pm *PeerManager) deletePeer(peerID peer.ID) {
	pm.peerStore.ClearAddrs(peerID)
	pm.routingTable.Remove(peerID)
	pm.lastUpdateTime.Store(time.Now().Unix())
}

// AddNeighbor starts a peer and adds it to the neighbor list.
func (pm *PeerManager) AddNeighbor(p *Peer) {
	p.Start()
	pm.storePeer(p.id, p.addr)

	pm.neighborMutex.Lock()
	defer pm.neighborMutex.Unlock()

	pm.neighbors[p.id] = p
}

// RemoveNeighbor stops a peer and removes it from the neighbor list.
func (pm *PeerManager) RemoveNeighbor(peerID peer.ID) {
	pm.deletePeer(peerID)

	pm.neighborMutex.Lock()
	defer pm.neighborMutex.Unlock()

	if peer, exist := pm.neighbors[peerID]; exist {
		peer.Stop()
		delete(pm.neighbors, peerID)
	}
}

// GetNeighbor returns the peer of the given peerID from the neighbor list.
func (pm *PeerManager) GetNeighbor(peerID peer.ID) *Peer {
	pm.neighborMutex.RLock()
	defer pm.neighborMutex.RUnlock()

	return pm.neighbors[peerID]
}

// NeighborCount returns the neighbor amount.
func (pm *PeerManager) NeighborCount() int {
	pm.neighborMutex.RLock()
	defer pm.neighborMutex.RUnlock()

	return len(pm.neighbors)
}

// DumpRoutingTable saves routing table in file.
func (pm *PeerManager) DumpRoutingTable() {
	file, err := os.Create(pm.config.RoutingFile)
	if err != nil {
		ilog.Error("create routing file failed. err=%v, path=%v", err, pm.config.RoutingFile)
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

// LoadRoutingTable reads routing table file and parses it.
func (pm *PeerManager) LoadRoutingTable() {
	if _, err := os.Stat(pm.config.RoutingFile); err != nil {
		if os.IsNotExist(err) {
			ilog.Info("no routing file. path=%v", pm.config.RoutingFile)
			return
		}
	}

	file, err := os.Open(pm.config.RoutingFile)
	if err != nil {
		ilog.Error("open routing file failed. err=%v, path=%v", err, pm.config.RoutingFile)
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
			ilog.Warn("parse multi addr failed. err=%v, line=%v", err, line)
			continue
		}
		pm.storePeer(peerID, addr)
	}
}

// syncRoutingTable broadcasts a routing table message. If the neighbor count is less than the threshold,
// it will pick the rest amount of peers from the routing table and sends query to them.
func (pm *PeerManager) syncRoutingTable() {
	pm.Broadcast(nil, RoutingTableQuery, UrgentMessage)
	neighborCount := pm.NeighborCount()
	if neighborCount >= maxNeighborCount {
		return
	}
	allPeerIDs := pm.routingTable.ListPeers()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	perm := r.Perm(len(allPeerIDs))

	for i, t := 0, 0; i < len(perm) && t < maxNeighborCount-neighborCount; i++ {
		peerID := allPeerIDs[perm[i]]
		if peerID == pm.host.ID() {
			continue
		}
		if pm.GetNeighbor(peerID) != nil {
			continue
		}
		stream, err := pm.host.NewStream(context.Background(), peerID, protocolID)
		if err != nil {
			ilog.Error("create stream failed. err=%v", err)
			pm.deletePeer(peerID)
			continue
		}
		pm.HandleStream(stream)
		pm.SendToPeer(peerID, nil, RoutingTableQuery, UrgentMessage)
		t++
	}
}

func (pm *PeerManager) parseSeeds() {
	for _, seed := range pm.config.SeedNodes {
		peerID, addr, err := parseMultiaddr(seed)
		if err != nil {
			ilog.Error("parse seed nodes error. seed=%s, err=%v", seed, err)
			continue
		}
		pm.storePeer(peerID, addr)
	}
}

// Broadcast sends message to all the neighbors.
func (pm *PeerManager) Broadcast(data []byte, typ MessageType, mp MessagePriority) {
	ilog.Info("broadcast message. type=%d", typ)
	msg := newP2PMessage(pm.config.ChainID, typ, pm.config.Version, defaultReservedFlag, data)

	pm.neighborMutex.RLock()
	defer pm.neighborMutex.RUnlock()

	for _, peer := range pm.neighbors {
		peer.SendMessage(msg, mp, true)
	}
}

// SendToPeer sends message to the specified peer.
func (pm *PeerManager) SendToPeer(peerID peer.ID, data []byte, typ MessageType, mp MessagePriority) {
	ilog.Info("send message to peer. type=%d, peerID=%s", typ, peerID.Pretty())
	msg := newP2PMessage(pm.config.ChainID, typ, pm.config.Version, defaultReservedFlag, data)

	peer := pm.GetNeighbor(peerID)
	if peer != nil {
		peer.SendMessage(msg, mp, false)
	}
}

// Register registers a message channel of the given types.
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

// Deregister deregisters a message channel of the given types.
func (pm *PeerManager) Deregister(id string, mTyps ...MessageType) {
	for _, typ := range mTyps {
		if m, exist := pm.subs.Load(typ); exist {
			m.(*sync.Map).Delete(id)
		}
	}
}

// handleRoutingTableQuery picks the nearest peers of the given peerID and sends the result to it.
func (pm *PeerManager) handleRoutingTableQuery(peerID peer.ID) {
	ilog.Info("handling routing table query. peerID=%s", peerID.Pretty())

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
		ilog.Error("json encode failed. err=%v, obj=%+v", err, peerInfo)
		return
	}
	pm.SendToPeer(peerID, bytes, RoutingTableResponse, UrgentMessage)
}

// handleRoutingTableResponse stores the peer information received.
func (pm *PeerManager) handleRoutingTableResponse(msg *p2pMessage) {
	ilog.Info("handling routing table response.")

	data, err := msg.data()
	if err != nil {
		ilog.Error("get message data failed. err=%v", err)
		return
	}
	peerInfos := make([]peerstore.PeerInfo, 0)
	err = json.Unmarshal(data, &peerInfos)
	if err != nil {
		ilog.Error("json decode failed. err=%v, str=%s", err, data)
		return
	}
	ilog.Info("receiving peer infos: %v", peerInfos)
	for _, peerInfo := range peerInfos {
		if len(peerInfo.Addrs) > 0 {
			pm.storePeer(peerInfo.ID, peerInfo.Addrs[0])
		}
	}
}

// HandleMessage handles messages according to its type.
func (pm *PeerManager) HandleMessage(msg *p2pMessage, peerID peer.ID) {
	data, err := msg.data()
	if err != nil {
		ilog.Error("get message data failed. err=%v", err)
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
					ilog.Error("sending incoming message failed. type=%d", msg.messageType())
				}
				return true
			})
		}
	}
}
