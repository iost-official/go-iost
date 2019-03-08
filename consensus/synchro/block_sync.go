package synchro

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/synchronizer/pb"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/patrickmn/go-cache"
)

const (
	requestCacheExpiration     = 10 * time.Second
	requestCachePurgeInterval  = 2 * time.Second
	responseCacheExpiration    = 5 * time.Minute
	responseCachePurgeInterval = 1 * time.Minute
)

// BlockMessage define a block from a neighbor node.
type BlockMessage struct {
	Blk     *block.Block
	P2PType p2p.MessageType
	From    string
}

type blockSync struct {
	p             p2p.Service
	requestCache  *cache.Cache
	responseCache *cache.Cache

	msgCh   chan p2p.IncomingMessage
	blockCh chan *BlockMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newBlockSync(p p2p.Service) *blockSync {
	b := &blockSync{
		p:             p,
		requestCache:  cache.New(requestCacheExpiration, requestCachePurgeInterval),
		responseCache: cache.New(responseCacheExpiration, responseCachePurgeInterval),

		msgCh:   p.Register("block from other nodes", p2p.SyncBlockResponse, p2p.NewBlock),
		blockCh: make(chan *BlockMessage, 1024),

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

func (b *blockSync) IncomingBlock() <-chan *BlockMessage {
	return b.blockCh
}

func (b *blockSync) RequestBlock(hash []byte, peerID p2p.PeerID) {
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

	b.p.SendToPeer(peerID, msg, p2p.SyncBlockRequest, p2p.UrgentMessage)
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

	blockMessage := &BlockMessage{
		Blk:     blk,
		P2PType: msg.Type(),
		From:    msg.From().Pretty(),
	}
	b.blockCh <- blockMessage
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
