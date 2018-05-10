package consensus_common

import (
	"github.com/iost-official/prototype/core/message"
	. "github.com/iost-official/prototype/network"
	"time"
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
			ReqBlockHeight,
		}})
	if err != nil {
		return nil
	}

	sync.blkSyncChain, err = sync.router.FilteredChan(Filter{
		WhiteList:  []message.Message{},
		BlackList:  []message.Message{},
		RejectType: []ReqType{},
		AcceptType: []ReqType{
			ReqDownloadBlock,
		}})
	if err != nil {
		return nil
	}
	return sync
}

//开始监听同步任务
func (sync *SyncImpl) StartListen() error {

	return nil
}

func (sync *SyncImpl) NeedSync() (bool, uint64, uint64) {
	height := sync.blockCache.ConfirmedLength()
	maxCachedHeight := sync.blockCache.MaxHeight()
	if height < maxCachedHeight-uint64(SyncNumber) {
		body := message.RequestHeight{
			LocalBlockHeight: height,
			NeedBlockHeight:  maxCachedHeight,
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

func (sync *SyncImpl) heightLoop() {

	for {
		req, ok := <-sync.heightChan
		if !ok {
			return
		}
		var rh message.RequestHeight

		rh.Decode(req.Body)

		chain := sync.blockCache.LongestChain()
		localLength := chain.Length()

		//本地链长度小于等于远端，忽略远端的同步链请求
		if localLength <= rh.LocalBlockHeight {
			continue
		}

		//回复当前块的高度
		hr :=message.ResponseHeight{BlockHeight:localLength}
		resMsg := message.Message{
			Time:time.Now().Unix(),
			From:req.To,
			To:req.From,
			ReqType:1,
			Body:hr.Encode(),
		}

		sync.router.Send(resMsg)
	}
}
