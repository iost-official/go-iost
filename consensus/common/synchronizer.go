package consensus_common

import (
	"time"

	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/log"
	. "github.com/iost-official/prototype/network"
)

var (
	SyncNumber        = 2  // 当本地链长度和网络中最新块相差SyncNumber时需要同步
	MaxDownloadNumber = 10 // 一次同步下载的最多块数
)

// Synchronizer 同步器接口
type Synchronizer interface {
	StartListen() error
	StopListen() error
	NeedSync(maxHeight uint64) (bool, uint64, uint64)
	SyncBlocks(startNumber uint64, endNumber uint64) error
}

// SyncImpl 同步器实现
type SyncImpl struct {
	blockCache    BlockCache
	router        Router
	maxSyncNumber uint64
	heightChan    chan message.Message
	blkSyncChan   chan message.Message
	exitSignal    chan struct{}

	log *log.Logger
}

// NewSynchronizer 新建同步器
// bc 块缓存, router 网络处理器
func NewSynchronizer(bc BlockCache, router Router) *SyncImpl {
	sync := &SyncImpl{
		blockCache: bc,
		router:     router,
	}
	if block.BChain != nil {
		sync.maxSyncNumber = block.BChain.Length() - 1
	}
	var err error
	sync.heightChan, err = sync.router.FilteredChan(Filter{
		AcceptType: []ReqType{
			ReqBlockHeight,
		}})
	if err != nil {
		return nil
	}

	sync.blkSyncChan, err = sync.router.FilteredChan(Filter{
		AcceptType: []ReqType{
			ReqDownloadBlock,
		}})
	if err != nil {
		return nil
	}

	sync.log, err = log.NewLogger("synchronizer.log")
	if err != nil {
		return nil
	}

	sync.log.NeedPrint = true

	return sync
}

// StartListen 开始监听同步任务
func (sync *SyncImpl) StartListen() error {
	go sync.requestBlockHeightLoop()
	go sync.requestBlockLoop()
	go sync.blockConfirmLoop()

	return nil
}

func (sync *SyncImpl) StopListen() error {
	close(sync.heightChan)
	close(sync.blkSyncChan)
	close(sync.blockCache.BlockConfirmChan())
	close(sync.exitSignal)
	return nil
}

func max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

// NeedSync 判断是否需要同步
// netHeight 当前网络收到的无法上链的块号
func (sync *SyncImpl) NeedSync(netHeight uint64) (bool, uint64, uint64) {
	//height := sync.blockCache.LongestChain().Length() - 1
	if netHeight > sync.maxSyncNumber+uint64(SyncNumber) {
		/*
			body := message.RequestHeight{
				LocalBlockHeight: height + 1,
				NeedBlockHeight:  netHeight,
			}
			heightReq := message.Message{
				ReqType: int32(ReqBlockHeight),
				Body:    body.Encode(),
			}
			sync.router.Broadcast(heightReq)
		*/
		return true, sync.maxSyncNumber + 1, netHeight
	}
	return false, 0, 0
}

// SyncBlocks 执行块同步操作
func (sync *SyncImpl) SyncBlocks(startNumber uint64, endNumber uint64) error {
	sync.maxSyncNumber = endNumber
	for endNumber > startNumber+uint64(MaxDownloadNumber) {
		sync.router.Download(startNumber, startNumber+uint64(MaxDownloadNumber))
		//TODO 等待所有区间里的块都收到
		time.Sleep(time.Second * 2)
		startNumber += uint64(MaxDownloadNumber + 1)
	}
	if startNumber <= endNumber {
		sync.router.Download(startNumber, endNumber)
	}
	return nil
}

func (sync *SyncImpl) requestBlockHeightLoop() {
	for {
		select {
		case req, ok := <-sync.heightChan:
			if !ok {
				return
			}
			var rh message.RequestHeight
			rh.Decode(req.Body)

			localLength := sync.blockCache.LongestChain().Length()

			//本地链长度小于等于远端，忽略远端的同步链请求
			if localLength <= rh.LocalBlockHeight {
				continue
			}
			sync.log.I("requset height - LocalBlockHeight: %v, NeedBlockHeight: %v", rh.LocalBlockHeight, rh.NeedBlockHeight)
			sync.log.I("local height: %v", localLength)

			//回复当前块的高度
			hr := message.ResponseHeight{BlockHeight: localLength}
			resMsg := message.Message{
				Time:    time.Now().Unix(),
				From:    req.To,
				To:      req.From,
				ReqType: int32(RecvBlockHeight),
				Body:    hr.Encode(),
			}
			sync.router.Send(resMsg)
		case <-sync.exitSignal:
			return
		}

	}
}

func (sync *SyncImpl) requestBlockLoop() {

	for {
		select {
		case req, ok := <-sync.blkSyncChan:
			if !ok {
				return
			}
			var rh message.RequestBlock
			rh.Decode(req.Body)

			chain := sync.blockCache.LongestChain()

			//todo 需要确定如何获取
			block := chain.GetBlockByNumber(rh.BlockNumber)
			if block == nil {
				continue
			}
			sync.log.I("requset block - BlockNumber: %v", rh.BlockNumber)
			sync.log.I("response block - BlockNumber: %v, witness: %v", block.Head.Number, block.Head.Witness)
			//回复当前块的高度
			resMsg := message.Message{
				Time:    time.Now().Unix(),
				From:    req.To,
				To:      req.From,
				ReqType: int32(ReqNewBlock), //todo 后补类型
				Body:    block.Encode(),
			}
			sync.router.Send(resMsg)
		case <-sync.exitSignal:
			return
		}

	}
}

func (sync *SyncImpl) blockConfirmLoop() {
	for {
		select {
		case num, ok := <-sync.blockCache.BlockConfirmChan():
			if !ok {
				return
			}
			sync.router.CancelDownload(num, num)
		case <-sync.exitSignal:
			return
		}

	}
}
