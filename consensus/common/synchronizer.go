package consensus_common

import (
	"github.com/iost-official/prototype/core/message"
	. "github.com/iost-official/prototype/network"
	"time"
)

type Synchronizer interface {
	StartListen() error
	RunSync() error
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

func (sync *SyncImpl) RunSync() error {
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
