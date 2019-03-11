package synchro

import (
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/consensus/synchro/pb"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

const (
	maxSyncRange = 1000
)

// Sync is the synchronizer of blockchain.
// It includes requestHandler, heightSync, blockhashSync, blockSync.
type Sync struct {
	p      p2p.Service
	bCache blockcache.BlockCache
	bChain block.Chain

	handler       *requestHandler
	heightSync    *heightSync
	blockhashSync *blockHashSync
	blockSync     *blockSync

	quitCh chan struct{}
	done   *sync.WaitGroup
}

// New will return a new synchronizer of blockchain.
func New(p p2p.Service, bCache blockcache.BlockCache, bChain block.Chain) *Sync {
	sync := &Sync{
		p:      p,
		bCache: bCache,
		bChain: bChain,

		handler:       newRequestHandler(p, bCache, bChain),
		heightSync:    newHeightSync(p),
		blockhashSync: newBlockHashSync(p),
		blockSync:     newBlockSync(p),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	sync.done.Add(4)
	go sync.heightSyncController()
	go sync.blockhashSyncController()
	go sync.blockSyncController()
	go sync.metricsController()

	return sync
}

// Close will close the synchronizer of blockchain.
func (s *Sync) Close() {
	s.handler.Close()
	s.heightSync.Close()
	s.blockhashSync.Close()
	s.blockSync.Close()

	close(s.quitCh)
	s.done.Wait()
	ilog.Infof("Stopped sync.")
}

// IncomingBlock will return the blocks from other nodes.
// Including passive request and active broadcast.
func (s *Sync) IncomingBlock() <-chan *BlockMessage {
	return s.blockSync.IncomingBlock()
}

// NeighborHeight will return the median of the head height of the neighbor nodes.
// If the number of neighbor nodes is less than leastNeighborNumber, return -1.
func (s *Sync) NeighborHeight() int64 {
	return s.heightSync.NeighborHeight()
}

func (s *Sync) doHeightSync() {
	syncHeight := &msgpb.SyncHeight{
		Height: s.bCache.Head().Head.Number,
		Time:   time.Now().Unix(),
	}
	msg, err := proto.Marshal(syncHeight)
	if err != nil {
		ilog.Errorf("Marshal sync height message failed: %v", err)
		return
	}
	s.p.Broadcast(msg, p2p.SyncHeight, p2p.UrgentMessage)
}

func (s *Sync) heightSyncController() {
	for {
		select {
		case <-time.After(1 * time.Second):
			s.doHeightSync()
		case <-s.quitCh:
			s.done.Done()
			return
		}
	}
}

func (s *Sync) doBlockhashSync() {
	now := time.Now().UnixNano()
	defer func() {
		blockHashSyncTimeGauge.Set(float64(time.Now().UnixNano()-now), nil)
	}()

	start := s.bCache.LinkedRoot().Head.Number + 1
	end := s.heightSync.NeighborHeight()
	if start > end {
		return
	}
	if end-start+1 > maxSyncRange {
		end = start + maxSyncRange - 1
	}

	s.blockhashSync.RequestBlockHash(start, end)
}

func (s *Sync) blockhashSyncController() {
	for {
		select {
		case <-time.After(2 * time.Second):
			s.doBlockhashSync()
		case <-s.quitCh:
			s.done.Done()
			return
		}
	}
}

func (s *Sync) doBlockSync() {
	now := time.Now().UnixNano()
	defer func() {
		blockSyncTimeGauge.Set(float64(time.Now().UnixNano()-now), nil)
	}()

	start := s.bCache.LinkedRoot().Head.Number + 1
	end := s.heightSync.NeighborHeight()
	if start > end {
		return
	}
	if end-start+1 > maxSyncRange {
		end = start + maxSyncRange - 1
	}

	ilog.Infof("Syncing block in [%v %v]...", start, end)
	for blockHash := range s.blockhashSync.NeighborBlockHashs(start, end) {
		if block, err := s.bCache.GetBlockByHash(blockHash.Hash); err == nil && block != nil {
			continue
		}

		rand.Seed(time.Now().UnixNano())
		peerID := blockHash.PeerID[rand.Int()%len(blockHash.PeerID)]
		s.blockSync.RequestBlock(blockHash.Hash, peerID)
	}
}

func (s *Sync) blockSyncController() {
	for {
		select {
		case <-time.After(2 * time.Second):
			s.doBlockSync()
		case <-s.quitCh:
			s.done.Done()
			return
		}
	}
}

func (s *Sync) metricsController() {
	for {
		select {
		case <-time.After(2 * time.Second):
			neighborHeightGauge.Set(float64(s.heightSync.NeighborHeight()), nil)
			incomingBlockBufferGauge.Set(float64(len(s.blockSync.IncomingBlock())), nil)
		case <-s.quitCh:
			s.done.Done()
			return
		}
	}
}
