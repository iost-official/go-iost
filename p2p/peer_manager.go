package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
)

var (
	dumpRoutingTableInterval = 5 * time.Minute
	syncRoutingTableInterval = 30 * time.Second
	metricsStatInterval      = 3 * time.Second
)

const (
	maxNeighborCount  = 32
	bucketSize        = 20
	peerResponseCount = 20

	incomingMsgChanSize = 1024

	routingTableFile = "routing.table"
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
	config         *common.P2PConfig
	routingTable   *kbucket.RoutingTable
	peerStore      peerstore.Peerstore
	lastUpdateTime atomic.Int64

	wg *sync.WaitGroup
}

// NewPeerManager returns a new instance of PeerManager struct.
func NewPeerManager(host host.Host, config *common.P2PConfig) *PeerManager {
	routingTable := kbucket.NewRoutingTable(bucketSize, kbucket.ConvertPeerID(host.ID()), time.Second, host.Peerstore())
	return &PeerManager{
		neighbors:    make(map[peer.ID]*Peer),
		subs:         new(sync.Map),
		quitCh:       make(chan struct{}),
		routingTable: routingTable,
		host:         host,
		config:       config,
		peerStore:    host.Peerstore(),
		wg:           new(sync.WaitGroup),
	}
}

// Start starts peer manager's job.
func (pm *PeerManager) Start() {
	pm.parseSeeds()
	pm.LoadRoutingTable()
	pm.syncRoutingTable()

	go pm.dumpRoutingTableLoop()
	go pm.syncRoutingTableLoop()
	go pm.metricsStatLoop()

}

// Stop stops peer manager's loop.
func (pm *PeerManager) Stop() {
	close(pm.quitCh)
	pm.wg.Wait()
}

// HandleStream handles the incoming stream.
//
// It checks whether the remote peer already exists.
// If the peer is new and the neighbor count doesn't reach the threshold, it adds the peer into the neighbor list.
// If peer already exits, just add the stream to the peer.
// In other cases, reset the stream.
func (pm *PeerManager) HandleStream(s libnet.Stream) {
	remotePID := s.Conn().RemotePeer()
	ilog.Infof("handle new stream. pid=%s, addr=%v", remotePID.Pretty(), s.Conn().RemoteMultiaddr())

	peer := pm.GetNeighbor(remotePID)
	if peer == nil {
		if pm.NeighborCount() >= maxNeighborCount {
			ilog.Debugf("reset stream. remoteID=%v, addr=%v", remotePID.Pretty(), s.Conn().RemoteMultiaddr())
			s.Conn().Close()
			return
		}
		pm.AddNeighbor(NewPeer(s, pm))
		return
	}

	err := peer.AddStream(s)
	if err != nil {
		ilog.Infof("add stream failed. err=%v, pid=%s", err, remotePID.Pretty())
		s.Reset()
		return
	}
}

func (pm *PeerManager) dumpRoutingTableLoop() {
	pm.wg.Add(1)
	var lastSaveTime int64
	dumpRoutingTableTicker := time.NewTimer(dumpRoutingTableInterval)
	for {
		select {
		case <-pm.quitCh:
			pm.wg.Done()
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
	pm.wg.Add(1)
	syncRoutingTableTicker := time.NewTimer(syncRoutingTableInterval)
	for {
		select {
		case <-pm.quitCh:
			pm.wg.Done()
			return
		case <-syncRoutingTableTicker.C:
			ilog.Infof("start sync routing table.")
			pm.syncRoutingTable()
			syncRoutingTableTicker.Reset(syncRoutingTableInterval)
		}
	}
}

func (pm *PeerManager) metricsStatLoop() {
	pm.wg.Add(1)
	metricsStatTicker := time.NewTimer(metricsStatInterval)
	for {
		select {
		case <-pm.quitCh:
			pm.wg.Done()
			return
		case <-metricsStatTicker.C:
			neighborCountGauge.Set(float64(pm.NeighborCount()), nil)
			routingCountGauge.Set(float64(pm.routingTable.Size()), nil)

			metricsStatTicker.Reset(metricsStatInterval)
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

// RestartNeighbor cuts off a neighbor's connection and reconnects it.
func (pm *PeerManager) RestartNeighbor(peerID peer.ID) {
	ilog.Debugf("restart neighbor, peerID=%s", peerID.Pretty())
	pm.RemoveNeighbor(peerID)

	stream, err := pm.host.NewStream(context.Background(), peerID, protocolID)
	if err != nil {
		ilog.Errorf("create stream failed. err=%v", err)
		return
	}
	pm.HandleStream(stream)

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
	file, err := os.Create(filepath.Join(pm.config.DataPath, routingTableFile))
	if err != nil {
		ilog.Errorf("create routing file failed. err=%v, path=%v", err, pm.config.DataPath)
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
	routingFile := filepath.Join(pm.config.DataPath + routingTableFile)
	if _, err := os.Stat(routingFile); err != nil {
		if os.IsNotExist(err) {
			ilog.Infof("no routing file. file=%v", routingFile)
			return
		}
	}

	file, err := os.Open(routingFile)
	if err != nil {
		ilog.Errorf("open routing file failed. err=%v, file=%v", err, routingFile)
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
		peerID, addr, err := parseMultiaddr(line[:len(line)-1])
		if err != nil {
			ilog.Warnf("parse multi addr failed. err=%v, line=%v", err, line)
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
			ilog.Errorf("create stream failed. err=%v", err)
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
			ilog.Errorf("parse seed nodes error. seed=%s, err=%v", seed, err)
			continue
		}
		pm.storePeer(peerID, addr)
	}
}

// Broadcast sends message to all the neighbors.
func (pm *PeerManager) Broadcast(data []byte, typ MessageType, mp MessagePriority) {
	/* if typ == PublishTxRequest { */
	// return
	/* } */
	if typ == NewBlock || typ == NewBlockHash || typ == SyncBlockHashRequest {
		ilog.Infof("broadcast message. type=%s", typ)
	}
	msg := newP2PMessage(pm.config.ChainID, typ, pm.config.Version, defaultReservedFlag, data)

	pm.neighborMutex.RLock()
	defer pm.neighborMutex.RUnlock()

	for _, peer := range pm.neighbors {
		peer.SendMessage(msg, mp, true)
	}
}

// SendToPeer sends message to the specified peer.
func (pm *PeerManager) SendToPeer(peerID peer.ID, data []byte, typ MessageType, mp MessagePriority) {
	if typ == NewBlock || typ == NewBlockRequest || typ == SyncBlockHashResponse ||
		typ == SyncBlockRequest || typ == SyncBlockResponse {
		ilog.Infof("send message to peer. type=%s, peerID=%s", typ, peerID.Pretty())
	}
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
	ilog.Infof("handling routing table query. peerID=%s", peerID.Pretty())

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
		ilog.Errorf("json encode failed. err=%v, obj=%+v", err, peerInfo)
		return
	}
	pm.SendToPeer(peerID, bytes, RoutingTableResponse, UrgentMessage)
}

// handleRoutingTableResponse stores the peer information received.
func (pm *PeerManager) handleRoutingTableResponse(msg *p2pMessage) {
	ilog.Infof("handling routing table response.")

	data, err := msg.data()
	if err != nil {
		ilog.Errorf("get message data failed. err=%v", err)
		return
	}
	peerInfos := make([]peerstore.PeerInfo, 0)
	err = json.Unmarshal(data, &peerInfos)
	if err != nil {
		ilog.Errorf("json decode failed. err=%v, str=%s", err, data)
		return
	}
	ilog.Infof("receiving peer infos: %v", peerInfos)
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
		ilog.Errorf("get message data failed. err=%v", err)
		return
	}
	if msg.messageType() != PublishTxRequest && msg.messageType() != SyncHeight {
		ilog.Infof("receiving message. type=%s", msg.messageType())
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
					ilog.Errorf("sending incoming message failed. type=%s", msg.messageType())
				}
				return true
			})
		}
	}
}
