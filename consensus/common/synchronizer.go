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
	SyncNumber                    = 2
	MaxBlockHashQueryNumber       = 10
	RetryTime                     = 5
	blockDownloadTimeout    int64 = 10
	cleanInterval                 = 5 * time.Second
	MaxAcceptableLength     int64 = 100
)

type Synchronizer interface {
	StartListen() error
	StopListen() error
	NeedSync(maxHeight uint64) (bool, uint64, uint64)
	SyncBlocks(startNumber uint64, endNumber uint64) error
	BlockConfirmed(num int64)
}

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

func (sync *SyncImpl) StartListen() error {
	go sync.requestBlockLoop()
	go sync.retryDownloadLoop()
	go sync.handleHashQuery()
	go sync.handleHashResp()
	go sync.recentAskedBlocksClean()
	return nil
}

func (sync *SyncImpl) StopListen() error {
	close(sync.blkSyncChan)
	close(sync.exitSignal)
	close(sync.blkHashQueryChan)
	close(sync.blkHashRespChan)
	return nil
}

func (sync *SyncImpl) NeedSync(netHeight uint64) (bool, uint64, uint64) {
	height := sync.blockCache.ConfirmedLength() - 1
	if netHeight > height+uint64(SyncNumber) {
		return true, height + 1, netHeight
	}

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
	if num > 0 {
		return true, height + 1, netHeight
	}

	return false, 0, 0
}

func (sync *SyncImpl) SyncBlocks(startNumber uint64, endNumber uint64) error {
	var syncNum int
	for endNumber > startNumber+uint64(MaxBlockHashQueryNumber)-1 {
		need := false
		for i := startNumber; i < startNumber+uint64(MaxBlockHashQueryNumber); i++ {
			_, ok := sync.requestMap.LoadOrStore(i, true)
			if !ok {
				need = true
			}
		}
		if need {
			syncNum++
			sync.router.QueryBlockHash(startNumber, startNumber+uint64(MaxBlockHashQueryNumber)-1)
		}
		startNumber += uint64(MaxBlockHashQueryNumber)
		if syncNum%10 == 0 {
			time.Sleep(time.Second)
		}
	}
	if startNumber <= endNumber {
		need := false
		for i := startNumber; i < endNumber; i++ {
			_, ok := sync.requestMap.LoadOrStore(i, true)
			if !ok {
				need = true
			}
		}
		if need {
			sync.router.QueryBlockHash(startNumber, endNumber)
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

			chain := sync.blockCache.BlockChain()
			var b []byte
			if rh.BlockNumber < chain.Length() {
				b, err = chain.GetBlockByteByHash(rh.BlockHash)
				if err != nil {
					log.Log.E("Database error: block empty %v", rh.BlockNumber)
					continue
				}
			} else {
				block, err := sync.blockCache.FindBlockInCache(rh.BlockHash)
				if err != nil {
					log.Log.E("Block not in cache: %v", rh.BlockNumber)
					continue
				}
				b = block.Encode()
			}

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
			delList := make([]uint64, 0)
			sync.requestMap.Range(func(k, v interface{}) bool {
				num, ok := k.(uint64)
				if !ok {
					return false
				}
				if num < sync.blockCache.ConfirmedLength() {
					delList = append(delList, num)
				} else {
					sync.router.QueryBlockHash(num, num)
				}
				return true
			})
			for _, num := range delList {
				sync.requestMap.Delete(num)
			}
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
				hash := chain.GetHashByNumber(i)
				if hash == nil {
					continue
				}
				blkHash := message.BlockHash{
					Height: i,
					Hash:   hash,
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
				if !sync.blockCache.CheckBlock(blkHash.Hash) {
					blkReq := &message.RequestBlock{
						BlockHash:   blkHash.Hash,
						BlockNumber: blkHash.Height,
					}
					if err != nil {
						sync.log.E("marshal BlockHashResponse failed:struct=%v, err=%v", blkReq, err)
						break
					}
					reqMsg := message.Message{
						Time:    time.Now().Unix(),
						From:    req.To,
						To:      req.From,
						ReqType: int32(ReqDownloadBlock),
						Body:    blkReq.Encode(),
					}
					sync.router.Send(reqMsg)
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
