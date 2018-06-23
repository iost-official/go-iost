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
	MaxBlockHashQueryNumber       = 10
	RetryTime                     = 8
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
	confirmNumber     int
	heightChan        chan message.Message
	blkSyncChan       chan message.Message
	blkHashQueryChan  chan message.Message
	blkHashRespChan   chan message.Message
	exitSignal        chan struct{}
	requestMap        *sy.Map
	recentAskedBlocks *sy.Map

	log *log.Logger
}

// NewSynchronizer 新建同步器
// bc 块缓存, router 网络处理器
func NewSynchronizer(bc blockcache.BlockCache, router Router, confirmNumber int) *SyncImpl {
	sync := &SyncImpl{
		blockCache:        bc,
		router:            router,
		requestMap:        new(sy.Map),
		confirmNumber:     confirmNumber,
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
		return true, height + 1, netHeight
	}

	// 如果在2/3长度的未确认链中出现了两次同一个witness，强制同步区块
	bc := sync.blockCache.LongestChain()
	ter := bc.Iterator()
	witness := bc.Top().Head.Witness
	num := 0
	for i := 0; i < sync.confirmNumber; i++ {
		block := ter.Next()
		if block == nil {
			break
		}
		if witness == block.Head.Witness {
			num++
		}
	}
	// 强制同步
	if num > 0 {
		return true, height + 1, netHeight
	}

	return false, 0, 0
}

// SyncBlocks 执行块同步操作
func (sync *SyncImpl) SyncBlocks(startNumber uint64, endNumber uint64) error {
	for endNumber > startNumber+uint64(MaxBlockHashQueryNumber)-1 {
		sync.router.QueryBlockHash(startNumber, startNumber+uint64(MaxBlockHashQueryNumber)-1)
		for i := startNumber; i < startNumber+uint64(MaxBlockHashQueryNumber); i++ {
			sync.requestMap.LoadOrStore(i, true)
		}
		startNumber += uint64(MaxBlockHashQueryNumber)
	}
	if startNumber <= endNumber {
		sync.router.QueryBlockHash(startNumber, endNumber)
		for i := startNumber; i <= endNumber; i++ {
			sync.requestMap.LoadOrStore(i, true)
		}
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
			var err error

			var rh message.RequestBlock
			err = rh.Decode(req.Body)
			if err != nil {
				continue
			}

			chain := sync.blockCache.LongestChain()

			b := make([]byte, 0)
			if rh.BlockNumber < chain.Length() { // 加速block的获取，减少encode
				b, err = chain.GetBlockByteByNumber(rh.BlockNumber)
				if err != nil {
					log.Log.E("Database error: block empty %v", rh.BlockNumber)
					continue
				}
			} else {
				block := chain.GetBlockByNumber(rh.BlockNumber)
				if block == nil {
					log.Log.E("Cache block empty %v", rh.BlockNumber)
					continue
				}
				b = block.Encode()
			}

			//sync.log.I("requset block - BlockNumber: %v", rh.BlockNumber)
			//sync.log.I("response block - BlockNumber: %v, witness: %v", block.Head.Number, block.Head.Witness)
			//回复当前高度的块
			resMsg := message.Message{
				Time:    time.Now().Unix(),
				From:    req.To,
				To:      req.From,
				ReqType: int32(ReqSyncBlock),
				Body:    b,
			}
			sync.router.Send(resMsg)
		case <-sync.exitSignal:
			return
		}

	}
}

func (sync *SyncImpl) BlockConfirmed(num int64) {
	sync.requestMap.Delete(uint64(num))
}

func (sync *SyncImpl) retryDownloadLoop() {
	for {
		select {
		case <-time.After(time.Second * time.Duration(RetryTime)):
			sync.requestMap.Range(func(k, v interface{}) bool {
				num, ok := k.(uint64)
				if !ok {
					return false
				}
				sync.router.QueryBlockHash(num, num)
				return true
			})
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
				//sync.log.I("check hash:%s, height:%v", blkHash.Hash, blkHash.Height)
				if !sync.blockCache.CheckBlock(blkHash.Hash) {
					//sync.log.I("check hash success")
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
