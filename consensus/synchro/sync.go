package synchro

import (
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
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

	handler         *requestHandler
	rangeController *rangeController
	heightSync      *heightSync
	blockhashSync   *blockHashSync
	blockSync       *blockSync

	quitCh chan struct{}
	done   *sync.WaitGroup
}

// New will return a new synchronizer of blockchain.
func New(p p2p.Service, bCache blockcache.BlockCache, bChain block.Chain) *Sync {
	sync := &Sync{
		p:      p,
		bCache: bCache,
		bChain: bChain,

		handler:         newRequestHandler(p, bCache, bChain),
		rangeController: newRangeController(bCache),
		heightSync:      newHeightSync(p),
		blockhashSync:   newBlockHashSync(p),
		blockSync:       newBlockSync(p),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	sync.done.Add(5)
	go sync.syncHeightController()
	go sync.syncBlockhashController()
	go sync.syncBlockController()
	go sync.syncNewBlockController()
	go sync.metricsController()

	return sync
}

// Close will close the synchronizer of blockchain.
func (s *Sync) Close() {
	s.handler.Close()
	s.rangeController.Close()
	s.heightSync.Close()
	s.blockhashSync.Close()
	s.blockSync.Close()

	close(s.quitCh)
	s.done.Wait()
	ilog.Infof("Stopped sync.")
}

// IncomingBlock will return the blocks from other nodes.
// Including passive request and active broadcast.
func (s *Sync) IncomingBlock() <-chan *block.Block {
	return s.blockSync.IncomingBlock()
}

// IsCatchingUp will return whether it is catching up with other nodes.
func (s *Sync) IsCatchingUp() bool {
	return s.bCache.Head().Head.Number+120 < s.heightSync.NeighborHeight()
}

// BroadcastBlockInfo will broadcast new block information to neighbor nodes.
func (s *Sync) BroadcastBlockInfo(block *block.Block) {
	// The block.Head.Number may not be used.
	blockInfo := &msgpb.BlockInfo{
		Number: block.Head.Number,
		Hash:   block.HeadHash(),
	}
	msg, err := proto.Marshal(blockInfo)
	if err != nil {
		ilog.Errorf("Marshal sync height message failed: %v", err)
		return
	}
	s.p.Broadcast(msg, p2p.NewBlockHash, p2p.UrgentMessage)
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

func (s *Sync) syncHeightController() {
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

	start, end := s.rangeController.SyncRange()
	nHeight := s.heightSync.NeighborHeight()
	if nHeight < end {
		end = nHeight
	}
	if start > end {
		return
	}

	s.blockhashSync.RequestBlockHash(start, end)
}

func (s *Sync) syncBlockhashController() {
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

	start, end := s.rangeController.SyncRange()
	nHeight := s.heightSync.NeighborHeight()
	if nHeight < end {
		end = nHeight
	}
	if start > end {
		return
	}

	ilog.Infof("Syncing block in [%v %v]...", start, end)
	for blockHash := range s.blockhashSync.NeighborBlockHashs(start, end) {
		if block, err := s.bCache.GetBlockByHash(blockHash.Hash); err == nil && block != nil {
			continue
		}

		rand.Seed(time.Now().UnixNano())
		peerID := blockHash.PeerID[rand.Int()%len(blockHash.PeerID)]
		s.blockSync.RequestBlock(blockHash.Hash, peerID, p2p.SyncBlockRequest)
	}
}

func (s *Sync) syncBlockController() {
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

func (s *Sync) doNewBlockSync(blockHash *BlockHash) {
	if s.IsCatchingUp() {
		// Synchronous mode does not process new block.
		return
	}

	// May not need to judge number.
	lib := s.bCache.LinkedRoot().Head.Number
	head := s.bCache.Head().Head.Number
	if (blockHash.Number <= lib) || (blockHash.Number > head+1000) {
		ilog.Debugf("New block hash exceed range %v to %v.", lib, head+1000)
		return
	}

	_, err := s.bCache.Find(blockHash.Hash)
	if err == nil {
		ilog.Debugf("New block hash %v already exists.", common.Base58Encode(blockHash.Hash))
		return
	}

	// New block hash just have 0 number peer ID.
	s.blockSync.RequestBlock(blockHash.Hash, blockHash.PeerID[0], p2p.NewBlockRequest)
}

func (s *Sync) syncNewBlockController() {
	for {
		select {
		case blockHash := <-s.blockhashSync.NewBlockHashs():
			s.doNewBlockSync(blockHash)
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
