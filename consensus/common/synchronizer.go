package consensus_common

import (
	sy "sync"
	"time"

	"github.com/iost-official/prototype/core/blockcache"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/log"
	. "github.com/iost-official/prototype/network"
)

var (
	SyncNumber                    = 2 // 当本地链长度和网络中最新块相差SyncNumber时需要同步
	MaxDownloadNumber             = 3 // 一次同步下载的最多块数
	MaxBlockHashQueryNumber       = 10
	RetryTime                     = 8
	continuousBlockNumber         = 2 //一个节点连续生产2个块，强制同步区块
	blockDownloadTimeout    int64 = 10
	cleanInterval                 = 5 * time.Second
)

// Synchronizer 同步器接口
type Synchronizer interface {
	StartListen() error
	StopListen() error
	NeedSync(maxHeight uint64) (bool, uint64, uint64)
	SyncBlocks(startNumber uint64, endNumber uint64) error
	BlockConfirmed(num int64)
}

// SyncImpl 同步器实现
type SyncImpl struct {
	blockCache        blockcache.BlockCache
	router            Router
	maxSyncNumber     uint64
	heightChan        chan message.Message
	blkSyncChan       chan message.Message
	blkHashQueryChan  chan message.Message
	blkHashRespChan   chan message.Message
	exitSignal        chan struct{}
	requestMap        map[uint64]bool
	reqMapLock        sy.RWMutex
	recentAskedBlocks *sy.Map

	log *log.Logger
}

// NewSynchronizer 新建同步器
// bc 块缓存, router 网络处理器
func NewSynchronizer(bc blockcache.BlockCache, router Router) *SyncImpl {
	sync := &SyncImpl{
		blockCache:        bc,
		router:            router,
		requestMap:        make(map[uint64]bool),
		maxSyncNumber:     0,
		recentAskedBlocks: new(sy.Map),
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

	sync.blkHashQueryChan, err = sync.router.FilteredChan(Filter{
		AcceptType: []ReqType{
			BlockHashQuery,
		}})
	if err != nil {
		return nil
	}

	sync.blkHashRespChan, err = sync.router.FilteredChan(Filter{
		AcceptType: []ReqType{
			BlockHashResponse,
		}})
	if err != nil {
		return nil
	}

	sync.log, err = log.NewLogger("synchronizer.log")
	if err != nil {
		return nil
	}

	sync.log.NeedPrint = false

	sync.log.I("maxSyncNumber:%v", sync.maxSyncNumber)

	return sync
}

// StartListen 开始监听同步任务
func (sync *SyncImpl) StartListen() error {
	//go sync.requestBlockHeightLoop()
	go sync.requestBlockLoop()
	go sync.retryDownloadLoop()
	go sync.handleHashQuery()
	go sync.handleHashResp()
	go sync.recentAskedBlocksClean()
	return nil
}

func (sync *SyncImpl) StopListen() error {
	//close(sync.heightChan)
	close(sync.blkSyncChan)
	//close(sync.blockCache.BlockConfirmChan())
	close(sync.exitSignal)
	close(sync.blkHashQueryChan)
	close(sync.blkHashRespChan)
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
	height := sync.blockCache.ConfirmedLength() - 1
	if netHeight > height+uint64(SyncNumber) {
		return true, max(sync.maxSyncNumber, height) + 1, netHeight
	}

	// 如果生产两个连续的块，强制同步区块，避免所有节点长度相同
	bc := sync.blockCache.LongestChain()
	ter := bc.Iterator()
	var witness string
	var i int
	for i = 0; i < continuousBlockNumber; i++ {
		block := ter.Next()
		if block == nil {
			break
		}

		if i == 0 {
			witness = block.Head.Witness
			continue
		}

		if witness != block.Head.Witness {
			break
		}
	}

	// 强制同步
	if i == continuousBlockNumber {
		return true, max(sync.maxSyncNumber, sync.blockCache.ConfirmedLength()-1) + 1, netHeight
	}

	return false, 0, 0
}

// SyncBlocks 执行块同步操作
func (sync *SyncImpl) SyncBlocks(startNumber uint64, endNumber uint64) error {
	sync.maxSyncNumber = max(sync.maxSyncNumber, endNumber)
	for endNumber > startNumber+uint64(MaxBlockHashQueryNumber)-1 {
		sync.router.QueryBlockHash(startNumber, startNumber+uint64(MaxBlockHashQueryNumber)-1)
		sync.reqMapLock.Lock()
		for i := startNumber; i < startNumber+uint64(MaxBlockHashQueryNumber); i++ {
			sync.requestMap[i] = true
		}
		sync.reqMapLock.Unlock()
		startNumber += uint64(MaxBlockHashQueryNumber)
	}
	if startNumber <= endNumber {
		sync.router.QueryBlockHash(startNumber, endNumber)
		sync.reqMapLock.Lock()
		for i := startNumber; i <= endNumber; i++ {
			sync.requestMap[i] = true
		}
		sync.reqMapLock.Unlock()
	}
	return nil
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
			//sync.log.I("requset block - BlockNumber: %v", rh.BlockNumber)
			//sync.log.I("response block - BlockNumber: %v, witness: %v", block.Head.Number, block.Head.Witness)
			//回复当前高度的块
			resMsg := message.Message{
				Time:    time.Now().Unix(),
				From:    req.To,
				To:      req.From,
				ReqType: int32(ReqNewBlock), //todo 后补类型
				Body:    block.Encode(),
			}
			////////////probe//////////////////
			log.Report(&log.MsgBlock{
				SubType:       "send",
				BlockHeadHash: block.HeadHash(),
				BlockNum:      block.Head.Number,
			})
			///////////////////////////////////
			sync.router.Send(resMsg)
		case <-sync.exitSignal:
			return
		}

	}
}

func (sync *SyncImpl) BlockConfirmed(num int64) {
	sync.reqMapLock.Lock()
	defer sync.reqMapLock.Unlock()
	delete(sync.requestMap, uint64(num))
}

func (sync *SyncImpl) retryDownloadLoop() {
	for {
		select {
		case <-time.After(time.Second * time.Duration(RetryTime)):
			sync.reqMapLock.RLock()
			for num, _ := range sync.requestMap {
				sync.router.QueryBlockHash(num, num)
			}
			sync.reqMapLock.RUnlock()
		case <-sync.exitSignal:
			return
		}
	}
}

func (sync *SyncImpl) handleHashQuery() {
	for {
		select {
		case req, ok := <-sync.blkHashQueryChan:
			if !ok {
				break
			}
			var rh message.BlockHashQuery
			_, err := rh.Unmarshal(req.Body)
			if err != nil {
				sync.log.E("unmarshal BlockHashQuery failed:%v", err)
				break
			}

			if rh.End < rh.Start {
				break
			}

			chain := sync.blockCache.LongestChain()

			resp := &message.BlockHashResponse{
				BlockHashes: make([]message.BlockHash, 0, rh.End-rh.Start+1),
			}
			for i := rh.Start; i <= rh.End; i++ {
				block := chain.GetBlockByNumber(i)
				if block == nil {
					continue
				}
				blkHash := message.BlockHash{
					Height: i,
					Hash:   block.HeadHash(),
				}
				resp.BlockHashes = append(resp.BlockHashes, blkHash)
			}
			if len(resp.BlockHashes) == 0 {
				break
			}
			bytes, err := resp.Marshal(nil)
			if err != nil {
				sync.log.E("marshal BlockHashResponse failed:struct=%v, err=%v", resp, err)
				break
			}
			resMsg := message.Message{
				Time:    time.Now().Unix(),
				From:    req.To,
				To:      req.From,
				ReqType: int32(BlockHashResponse),
				Body:    bytes,
			}
			sync.router.Send(resMsg)
		case <-sync.exitSignal:
			return
		}

	}
}

func (sync *SyncImpl) handleHashResp() {

	for {
		select {
		case req, ok := <-sync.blkHashRespChan:
			if !ok {
				break
			}
			var rh message.BlockHashResponse
			_, err := rh.Unmarshal(req.Body)
			if err != nil {
				sync.log.E("unmarshal BlockHashResponse failed:%v", err)
				break
			}

			sync.log.I("receive block hashes: len=%v", len(rh.BlockHashes))
			for _, blkHash := range rh.BlockHashes {
				if _, exist := sync.recentAskedBlocks.Load(string(blkHash.Hash)); exist {
					continue
				}
				// TODO: 判断本地是否有这个区块
				sync.log.I("chech hash:%s, height:%v", blkHash.Hash, blkHash.Height)
				if !sync.blockCache.CheckBlock(blkHash.Hash) {
					sync.log.I("check hash success")
					sync.router.AskABlock(blkHash.Height, req.From)
					sync.recentAskedBlocks.Store(string(blkHash.Hash), time.Now().Unix())
				}
			}

		case <-sync.exitSignal:
			return
		}
	}
}

func (sync *SyncImpl) recentAskedBlocksClean() {
	for {
		select {
		case <-time.After(cleanInterval):
			sync.recentAskedBlocks.Range(func(k, v interface{}) bool {
				t, ok := v.(int64)
				if !ok {
					sync.recentAskedBlocks.Delete(k)
					return true
				}
				if time.Now().Unix()-t > blockDownloadTimeout {
					sync.recentAskedBlocks.Delete(k)
				}
				return true
			})
		case <-sync.exitSignal:
			return
		}
	}
}
