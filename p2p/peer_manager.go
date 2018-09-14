package p2p

import (
	"bufio"
	"context"
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
	p2pb "github.com/iost-official/Go-IOS-Protocol/p2p/pb"
)

var (
	dumpRoutingTableInterval = 5 * time.Minute
	syncRoutingTableInterval = 30 * time.Second
	metricsStatInterval      = 3 * time.Second
	findBPInterval           = 2 * time.Second
)

const (
	maxNeighborCount  = 32
	bucketSize        = 20
	peerResponseCount = 20
	maxPeerQuery      = 30

	incomingMsgChanSize = 102400

	routingTableFile = "routing.table"
)

// PeerManager manages all peers we connect directily.
//
// PeerManager's jobs are:
//   * holding a certain amount of peers.
//   * handling messages according to its type.
//   * discovering peers and maintaining routing table.
type PeerManager struct {
	neighbors     *sync.Map // map[peer.ID]*Peer
	neighborCount int

	subs   *sync.Map //  map[MessageType]map[string]chan IncomingMessage
	quitCh chan struct{}

	host           host.Host
	config         *common.P2PConfig
	routingTable   *kbucket.RoutingTable
	peerStore      peerstore.Peerstore
	lastUpdateTime atomic.Int64

	wg *sync.WaitGroup

	bpIDs   []peer.ID
	bpMutex sync.RWMutex
}

// NewPeerManager returns a new instance of PeerManager struct.
func NewPeerManager(host host.Host, config *common.P2PConfig) *PeerManager {
	routingTable := kbucket.NewRoutingTable(bucketSize, kbucket.ConvertPeerID(host.ID()), time.Second, host.Peerstore())
	return &PeerManager{
		neighbors:    new(sync.Map),
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
	pm.routingQuery([]string{pm.host.ID().Pretty()})

	go pm.dumpRoutingTableLoop()
	go pm.syncRoutingTableLoop()
	go pm.metricsStatLoop()
	go pm.findBPLoop()

}

// Stop stops peer manager's loop.
func (pm *PeerManager) Stop() {
	close(pm.quitCh)
	pm.wg.Wait()
}

func (pm *PeerManager) isBP(id peer.ID) bool {
	for _, bp := range pm.getBPs() {
		if bp == id {
			return true
		}
	}
	return false
}

func (pm *PeerManager) setBPs(ids []string) {
	peerIDs := make([]peer.ID, 0, len(ids))
	for _, id := range ids {
		peerID, err := peer.IDB58Decode(id)
		if err != nil {
			ilog.Warnf("decode peerID failed. err=%v, id=%v", err, id)
			continue
		}
		peerIDs = append(peerIDs, peerID)
	}
	pm.bpMutex.Lock()
	pm.bpIDs = peerIDs
	pm.bpMutex.Unlock()
}

func (pm *PeerManager) getBPs() []peer.ID {
	pm.bpMutex.RLock()
	defer pm.bpMutex.RUnlock()

	return pm.bpIDs
}

func (pm *PeerManager) findBPLoop() {
	pm.wg.Add(1)
	for {
		select {
		case <-pm.quitCh:
			pm.wg.Done()
			return
		case <-time.After(findBPInterval):
			unknownBPs := make([]string, 0)
			for _, id := range pm.getBPs() {
				if len(pm.peerStore.Addrs(id)) == 0 {
					unknownBPs = append(unknownBPs, id.Pretty())
				}
			}
			pm.routingQuery(unknownBPs)
			pm.connectBPs()
		}
	}
}

func (pm *PeerManager) connectBPs() {
	for _, bpID := range pm.getBPs() {
		if pm.GetNeighbor(bpID) == nil && bpID != pm.host.ID() && len(pm.peerStore.Addrs(bpID)) > 0 {
			stream, err := pm.host.NewStream(context.Background(), bpID, protocolID)
			if err != nil {
				ilog.Errorf("create stream failed. pid=%s, err=%v", bpID.Pretty(), err)
				continue
			}
			pm.HandleStream(stream)
		}
	}
}

// ConnectBPs makes the local host connected to the block producers directly.
func (pm *PeerManager) ConnectBPs(ids []string) {
	if len(ids) == 0 {
		return
	}
	pm.setBPs(ids)
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
			if !pm.isBP(remotePID) {
				ilog.Debugf("neighbor count exceeds, close stream. remoteID=%v, addr=%v", remotePID.Pretty(), s.Conn().RemoteMultiaddr())
				s.Conn().Close()
				return
			}
			pm.kickNormalNeighbors()
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
	for {
		select {
		case <-pm.quitCh:
			pm.wg.Done()
			return
		case <-time.After(dumpRoutingTableInterval):
			if lastSaveTime < pm.lastUpdateTime.Load() {
				pm.DumpRoutingTable()
				lastSaveTime = time.Now().Unix()
			}
		}
	}
}

func (pm *PeerManager) syncRoutingTableLoop() {
	pm.wg.Add(1)
	for {
		select {
		case <-pm.quitCh:
			pm.wg.Done()
			return
		case <-time.After(syncRoutingTableInterval):
			ilog.Infof("start sync routing table.")
			pm.routingQuery([]string{pm.host.ID().Pretty()})
		}
	}
}

func (pm *PeerManager) metricsStatLoop() {
	pm.wg.Add(1)
	for {
		select {
		case <-pm.quitCh:
			pm.wg.Done()
			return
		case <-time.After(metricsStatInterval):
			neighborCountGauge.Set(float64(pm.NeighborCount()), nil)
			routingCountGauge.Set(float64(pm.routingTable.Size()), nil)
		}
	}

}

// storePeer stores peer information in peerStore and routingTable. It doesn't need lock since the
// peerStore.SetAddr and routingTable.Update function are thread safe.
func (pm *PeerManager) storePeer(peerID peer.ID, addrs []multiaddr.Multiaddr) {
	pm.peerStore.AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)
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
	pm.storePeer(p.id, []multiaddr.Multiaddr{p.addr})

	pm.neighbors.Store(p.id, p)
	pm.neighborCount++
}

// RemoveNeighbor stops a peer and removes it from the neighbor list.
func (pm *PeerManager) RemoveNeighbor(peerID peer.ID) {
	// pm.deletePeer(peerID)

	if peer, ok := pm.neighbors.Load(peerID); ok {
		peer.(*Peer).Stop()
		pm.neighbors.Delete(peerID)
		pm.neighborCount--
	}
}

// GetNeighbor returns the peer of the given peerID from the neighbor list.
func (pm *PeerManager) GetNeighbor(peerID peer.ID) *Peer {
	if peer, ok := pm.neighbors.Load(peerID); ok {
		return peer.(*Peer)
	}
	return nil
}

// NeighborCount returns the neighbor amount.
func (pm *PeerManager) NeighborCount() int {
	return pm.neighborCount
}

// kickNormalNeighbors removes neighbors that are not block producers.
func (pm *PeerManager) kickNormalNeighbors() {

	pm.neighbors.Range(func(k, v interface{}) bool {
		if pm.neighborCount < maxNeighborCount {
			return false
		}
		if !pm.isBP(k.(peer.ID)) {
			v.(*Peer).Stop()
			pm.neighbors.Delete(k)
			pm.neighborCount--
		}
		return true
	})
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
		pm.storePeer(peerID, []multiaddr.Multiaddr{addr})
	}
}

// routingQuery broadcasts a routing query message. If the neighbor count is less than the threshold,
// it will pick the rest amount of peers from the routing table and sends query to them.
func (pm *PeerManager) routingQuery(ids []string) {
	if len(ids) == 0 {
		return
	}
	query := p2pb.RoutingQuery{
		Ids: ids,
	}
	bytes, err := query.Marshal()
	if err != nil {
		ilog.Errorf("pb encode failed. err=%v, obj=%+v", err, query)
		return
	}

	pm.Broadcast(bytes, RoutingTableQuery, UrgentMessage)
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
			ilog.Errorf("create stream failed. pid=%s, err=%v", peerID.Pretty(), err)
			pm.deletePeer(peerID)
			continue
		}
		pm.HandleStream(stream)
		pm.SendToPeer(peerID, bytes, RoutingTableQuery, UrgentMessage)
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
		pm.storePeer(peerID, []multiaddr.Multiaddr{addr})
	}
}

// Broadcast sends message to all the neighbors.
func (pm *PeerManager) Broadcast(data []byte, typ MessageType, mp MessagePriority) {
	if typ == NewBlock || typ == NewBlockHash || typ == SyncBlockHashRequest {
		ilog.Infof("broadcast message. type=%s", typ)
	}
	msg := newP2PMessage(pm.config.ChainID, typ, pm.config.Version, defaultReservedFlag, data)

	pm.neighbors.Range(func(k, v interface{}) bool {
		v.(*Peer).SendMessage(msg, mp, true)
		return true
	})
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

// handleRoutingTableQuery picks the nearest peers of the given peerIDs and sends the result to it.
func (pm *PeerManager) handleRoutingTableQuery(msg *p2pMessage, peerID peer.ID) {
	ilog.Debug("handling routing table query.")
	data, _ := msg.data()

	query := &p2pb.RoutingQuery{}
	err := query.Unmarshal(data)
	if err != nil {
		ilog.Errorf("pb decode failed. err=%v, str=%s", err, data)
		return
	}

	queryIDs := query.GetIds()
	if len(queryIDs) > maxPeerQuery {
		queryIDs = queryIDs[:maxPeerQuery]
	}

	pidSet := make(map[peer.ID]struct{})
	for _, queryID := range queryIDs {
		pid, err := peer.IDB58Decode(queryID)
		if err != nil {
			ilog.Warnf("decode peerID failed. err=%v, id=%v", err, queryID)
			continue
		}
		peerIDs := pm.routingTable.NearestPeers(kbucket.ConvertPeerID(pid), peerResponseCount)
		for _, id := range peerIDs {
			pidSet[id] = struct{}{}
		}
	}

	resp := p2pb.RoutingResponse{}
	for pid := range pidSet {
		info := pm.peerStore.PeerInfo(pid)
		if len(info.Addrs) > 0 {
			peerInfo := &p2pb.PeerInfo{
				Id: info.ID.Pretty(),
			}
			for _, addr := range info.Addrs {
				peerInfo.Addrs = append(peerInfo.Addrs, addr.String())
			}
			resp.Peers = append(resp.Peers, peerInfo)
		}
	}
	if len(resp.Peers) == 0 {
		return
	}
	bytes, err := resp.Marshal()
	if err != nil {
		ilog.Errorf("pb encode failed. err=%v, obj=%+v", err, resp)
		return
	}
	pm.SendToPeer(peerID, bytes, RoutingTableResponse, UrgentMessage)
}

// handleRoutingTableResponse stores the peer information received.
func (pm *PeerManager) handleRoutingTableResponse(msg *p2pMessage) {
	ilog.Debug("handling routing table response.")

	data, _ := msg.data()

	resp := &p2pb.RoutingResponse{}
	err := resp.Unmarshal(data)
	if err != nil {
		ilog.Errorf("pb decode failed. err=%v, str=%s", err, data)
		return
	}
	ilog.Debugf("receiving peer infos: %v", resp)
	for _, peerInfo := range resp.Peers {
		if len(peerInfo.Addrs) > 0 {
			pid, err := peer.IDB58Decode(peerInfo.Id)
			if err != nil {
				ilog.Warnf("decode peerID failed. err=%v, id=%v", err, peerInfo.Id)
				continue
			}
			maddrs := make([]multiaddr.Multiaddr, 0, len(peerInfo.Addrs))
			for _, addr := range peerInfo.Addrs {
				a, err := multiaddr.NewMultiaddr(addr)
				if err != nil {
					ilog.Warnf("parse multiaddr failed. err=%v, addr=%v", err, addr)
					continue
				}
				maddrs = append(maddrs, a)
			}
			if len(maddrs) > 0 {
				pm.storePeer(pid, maddrs)
			}
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
		go pm.handleRoutingTableQuery(msg, peerID)
	case RoutingTableResponse:
		go pm.handleRoutingTableResponse(msg)
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

// NeighborStat dumps neighbors' status for debug.
func (pm *PeerManager) NeighborStat() map[string]interface{} {
	ret := make(map[string]interface{})

	pm.neighbors.Range(func(k, v interface{}) bool {
		ret[k.(peer.ID).Pretty()] = map[string]interface{}{
			"stream": v.(*Peer).streamCount,
		}
		return true
	})

	return ret
}
