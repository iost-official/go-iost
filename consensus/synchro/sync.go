package synchro

import (
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/chainbase"
	"github.com/iost-official/go-iost/common"
	msgpb "github.com/iost-official/go-iost/consensus/synchro/pb"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

const (
	maxSyncRange = 1000
)

// Sync is the synchronizer of blockchain.
// It includes requestHandler, heightSync, blockhashSync, blockSync.
type Sync struct {
	cBase   *chainbase.ChainBase
	p       p2p.Service
	blockCh chan *block.Block

	handler         *requestHandler
	rangeController *rangeController
	heightSync      *heightSync
	blockhashSync   *blockHashSync
	blockSync       *blockSync

	quitCh chan struct{}
	done   *sync.WaitGroup
}

// New will return a new synchronizer of blockchain.
func New(cBase *chainbase.ChainBase, p p2p.Service) *Sync {
	sync := &Sync{
		cBase:   cBase,
		p:       p,
		blockCh: make(chan *block.Block, 1024),

		handler:         newRequestHandler(cBase, p),
		rangeController: newRangeController(cBase),
		heightSync:      newHeightSync(p),
		blockhashSync:   newBlockHashSync(p),
		blockSync:       newBlockSync(p),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	sync.done.Add(6)
	go sync.syncHeightController()
	go sync.syncBlockhashController()
	go sync.syncBlockController()
	go sync.handleNewBlockHashController()
	go sync.handleBlockController()
	go sync.metricsController()

	return sync
}

// Close will close the synchronizer of blockchain.
func (s *Sync) Close() {
	close(s.quitCh)
	s.done.Wait()

	s.handler.Close()
	s.rangeController.Close()
	s.heightSync.Close()
	s.blockhashSync.Close()
	s.blockSync.Close()

	ilog.Infof("Stopped sync.")
}

// ValidBlock will return the valid blocks from other nodes.
// Including passive request and active broadcast.
func (s *Sync) ValidBlock() <-chan *block.Block {
	return s.blockCh
}

// IsCatchingUp will return whether it is catching up with other nodes.
func (s *Sync) IsCatchingUp() bool {
	return s.cBase.HeadBlock().Head.Number+120 < s.heightSync.NeighborHeight()
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

// BroadcastBlock will broadcast new block to neighbor nodes.
func (s *Sync) BroadcastBlock(block *block.Block) {
	msg, err := block.Encode()
	if err != nil {
		ilog.Errorf("Encode block failed: %v", err)
		return
	}
	s.p.Broadcast(msg, p2p.NewBlock, p2p.UrgentMessage)
}

func (s *Sync) doHeightSync() {
	syncHeight := &msgpb.SyncHeight{
		Height: s.cBase.HeadBlock().Head.Number,
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
	s.blockhashSync.RequestBlockHash(start, end)
}

func (s *Sync) syncBlockhashController() {
	//ilog.Debug("in syncBlockhashController")
	for {
		select {
		case <-time.After(2 * time.Second):
			s.doBlockhashSync()
		case <-s.quitCh:
			s.done.Done()
			return
		}
		//ilog.Debug("loop syncBlockhashController")
	}
}

func (s *Sync) doBlockSync() {
	now := time.Now().UnixNano()
	defer func() {
		blockSyncTimeGauge.Set(float64(time.Now().UnixNano()-now), nil)
	}()

	start, end := s.rangeController.SyncRange()
	ilog.Infof("Syncing block in [%v %v]...", start, end)
	for blockHash := range s.blockhashSync.NeighborBlockHashs(start, end) {
		_, err := s.cBase.BlockCache().GetBlockByHash(blockHash.Hash)
		if err == nil {
			ilog.Debugf("Block %v is existed, don't sync.", common.Base58Encode(blockHash.Hash))
			continue
		}

		rand.Seed(time.Now().UnixNano())
		peerID := blockHash.PeerID[rand.Int()%len(blockHash.PeerID)]
		//ilog.Debug("sync block ", blockHash.Number, " from ", peerID)
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
	lib := s.cBase.LIBlock().Head.Number
	head := s.cBase.HeadBlock().Head.Number
	if (blockHash.Number <= lib) || (blockHash.Number > head+1000) {
		ilog.Debugf("New block hash exceed range %v to %v.", lib, head+1000)
		return
	}

	_, err := s.cBase.BlockCache().GetBlockByHash(blockHash.Hash)
	if err == nil {
		ilog.Debugf("New block hash %v already exists.", common.Base58Encode(blockHash.Hash))
		return
	}

	// New block hash just have 0 number peer ID.
	s.blockSync.RequestBlock(blockHash.Hash, blockHash.PeerID[0], p2p.NewBlockRequest)
}

func (s *Sync) handleNewBlockHashController() {
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

func (s *Sync) doBlockFilter(block *block.Block) {
	head := s.cBase.HeadBlock().Head.Number
	lib := s.cBase.LIBlock().Head.Number
	if block.Head.Number > head+maxSyncRange {
		ilog.Debugf("Block number %v is %v higher than head number %v", block.Head.Number, maxSyncRange, head)
		return
	}
	if block.Head.Number <= lib {
		ilog.Debugf("Block number %v is lower than or equal to lib number %v", block.Head.Number, lib)
		return
	}

	s.blockCh <- block
}

func (s *Sync) handleBlockController() {
	for {
		select {
		case block := <-s.blockSync.IncomingBlock():
			s.doBlockFilter(block)
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
