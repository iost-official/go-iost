package synchro

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/synchro/pb"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/patrickmn/go-cache"
)

const (
	requestCacheExpiration     = 10 * time.Second
	requestCachePurgeInterval  = 1 * time.Minute
	responseCacheExpiration    = 10 * time.Second
	responseCachePurgeInterval = 1 * time.Minute
)

// blockSync is responsible for receiving neighbor's block and removing duplicate requests and responses.
type blockSync struct {
	p             p2p.Service
	requestCache  *cache.Cache
	responseCache *cache.Cache
	blockCh       chan *block.Block

	msgCh chan p2p.IncomingMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newBlockSync(p p2p.Service) *blockSync {
	b := &blockSync{
		p:             p,
		requestCache:  cache.New(requestCacheExpiration, requestCachePurgeInterval),
		responseCache: cache.New(responseCacheExpiration, responseCachePurgeInterval),
		blockCh:       make(chan *block.Block, 1024),

		msgCh: p.Register("block from other nodes", p2p.SyncBlockResponse, p2p.NewBlock),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	b.done.Add(1)
	go b.controller()

	return b
}

func (b *blockSync) Close() {
	close(b.quitCh)
	b.done.Wait()
	ilog.Infof("Stopped block sync.")
}

// IncomingBlock will return the blocks from other nodes.
func (b *blockSync) IncomingBlock() <-chan *block.Block {
	return b.blockCh
}

func (b *blockSync) RequestBlock(hash []byte, peerID p2p.PeerID, mtype p2p.MessageType) {
	// Filter duplicate requests in the short term
	_, found := b.requestCache.Get(string(hash))
	if found {
		ilog.Debugf("Discard the duplicate request block %v", common.Base58Encode(hash))
		return
	}
	b.requestCache.Set(string(hash), "", cache.DefaultExpiration)

	// Historical issues cause number to be useless.
	blockInfo := &msgpb.BlockInfo{
		Hash:   hash,
		Number: -1,
	}
	msg, err := proto.Marshal(blockInfo)
	if err != nil {
		ilog.Errorf("Marshal sync block message failed: %v", err)
		return
	}

	b.p.SendToPeer(peerID, msg, mtype, p2p.UrgentMessage)
}

func (b *blockSync) handleBlock(msg *p2p.IncomingMessage) {
	if (msg.Type() != p2p.SyncBlockResponse) && (msg.Type() != p2p.NewBlock) {
		ilog.Warnf("Expect the type %v and %v, but get a unexpected type %v", p2p.SyncBlockResponse, p2p.NewBlock, msg.Type())
		return
	}

	blk := &block.Block{}
	err := blk.Decode(msg.Data())
	if err != nil {
		ilog.Warnf("Decode block failed: %v", err)
		return
	}

	// Discard the most recently received duplicate block by hash
	_, found := b.responseCache.Get(string(blk.HeadHash()))
	if found {
		ilog.Debugf("Discard the duplicate received block %v", common.Base58Encode(blk.HeadHash()))
		return
	}
	b.responseCache.Set(string(blk.HeadHash()), "", cache.DefaultExpiration)

	ilog.Debugf("Received block %v from peer %v, num: %v", common.Base58Encode(blk.HeadHash()), msg.From().Pretty(), blk.Head.Number)

	b.blockCh <- blk
}

func (b *blockSync) controller() {
	for {
		select {
		case msg := <-b.msgCh:
			b.handleBlock(&msg)
		case <-b.quitCh:
			b.done.Done()
			return
		}
	}
}
