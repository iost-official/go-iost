package consensus_common

import (
	"github.com/iost-official/prototype/core/message"
	. "github.com/iost-official/prototype/network"
)

var (
	SyncNumber        = 10
	MaxDownloadNumber = 10
)

type Synchronizer interface {
	StartListen() error
	NeedSync() (bool, uint64, uint64)
	SyncBlocks(startNumber uint64, endNumber uint64) error
}

type SyncImpl struct {
	blockCache   BlockCache
	router       Router
	heightChan   chan message.Message
	blkSyncChain chan message.Message
}

func NewSynchronizer(bc BlockCache, router Router) *SyncImpl {
	sync := &SyncImpl{
		blockCache: bc,
		router:     router,
	}
	var err error
	sync.heightChan, err = sync.router.FilteredChan(Filter{
		WhiteList:  []message.Message{},
		BlackList:  []message.Message{},
		RejectType: []ReqType{},
		AcceptType: []ReqType{
			ReqHeight,
		}})
	if err != nil {
		return nil
	}

	sync.blkSyncChain, err = sync.router.FilteredChan(Filter{
		WhiteList:  []message.Message{},
		BlackList:  []message.Message{},
		RejectType: []ReqType{},
		AcceptType: []ReqType{
			ReqBlockSync,
		}})
	if err != nil {
		return nil
	}
	return sync
}

type HeightRequest struct {
	localHeight uint64
	needHeight  uint64
}

type HeightResponse struct {
	height uint64
}

type BlockRequest struct {
	number uint64
}

func (sync *SyncImpl) StartListen() error {
	return nil
}

func (sync *SyncImpl) NeedSync() (bool, uint64, uint64) {
	height := sync.blockCache.ConfirmedLength()
	maxCachedHeight := sync.blockCache.MaxHeight()
	if height < maxCachedHeight-uint64(SyncNumber) {
		body := HeightRequest{
			localHeight: height,
			needHeight:  maxCachedHeight,
		}
		heightReq := message.Message{
			From:    "",
			ReqType: int32(ReqBlockHeight),
			Body:    body.Encode(),
		}
		sync.router.Broadcast(heightReq)
		return true, height + 1, maxCachedHeight
	}
	return false, 0, 0
}

func (sync *SyncImpl) SyncBlocks(startNumber uint64, endNumber uint64) error {
	for endNumber-startNumber > uint64(MaxDownloadNumber) {
		sync.router.Download(startNumber, startNumber+uint64(MaxDownloadNumber))
		//TODO 等待所有区间里的块都收到
		startNumber += uint64(MaxDownloadNumber + 1)
	}
	sync.router.Download(startNumber, endNumber)
	return nil
}
