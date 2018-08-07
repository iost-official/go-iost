package consensus_common

import (
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/network"
)

var (
	SyncNumber                    = 2
	MaxBlockHashQueryNumber       = 10
	RetryTime                     = 5 * time.Second
	blockDownloadTimeout    int64 = 10
	MaxAcceptableLength     int64 = 100
)

type Synchronizer interface {
	Start() error
	Stop() error
	NeedSync(maxHeight uint64) (bool, uint64, uint64)
	SyncBlocks(startNumber uint64, endNumber uint64) error
	OnBlockConfirmed(hash, peerID string) error
}

type SyncImpl struct {
	router        Router
	dc            DownloadController
	confirmNumber int
	messageChan   chan message.Message
	exitSignal    chan struct{}

	log *log.Logger
}

func NewSynchronizer(bc blockcache.BlockCache, router Router, confirmNumber int) (*SyncImpl, error) {
	sync := &SyncImpl{
		blockCache: bc,
		router:     router,
		requestMap: new(sy.Map),
	}
	var err error
	sync.dc, err = NewDownloadController()
	if err != nil {
		return nil, err
	}

	sync.messageChan, err = sync.router.FilteredChan(Filter{
		AcceptType: []ReqType{
			network.ReqDownloadBlock,
			network.BlockHashQuery,
			network.BlockHashResponse,
		}})
	if err != nil {
		return nil, err
	}

	sync.log, err = log.NewLogger("synchronizer.log")
	if err != nil {
		return nil, err
	}

	sync.log.NeedPrint = false
	sync.exitSignal = make(chan struct{})

	return sync, nil
}

func (sy *SyncImpl) reqDownloadBlock(hash, peerID string) {
	blkReq := &message.RequestBlock{
		BlockHash: blkHash.Hash,
	}
	if err != nil {
		sy.log.E("marshal BlockHashResponse failed:struct=%v, err=%v", blkReq, err)
		break
	}
	reqMsg := message.Message{
		Time:    time.Now().Unix(),
		To:      peerID,
		ReqType: int32(ReqDownloadBlock),
		Body:    blkReq.Encode(),
	}
	sy.router.Send(reqMsg)
}

func (sy *SyncImpl) Start() error {
	go sy.dc.DownloadLoop(sy.reqDownloadBlock)
	go sy.messageLoop()
	go sy.retryDownloadLoop()
	return nil
}

func (sy *SyncImpl) Stop() error {
	sy.dc.Stop()
	close(sy.exitSignal)
	return nil
}

func (sy *SyncImpl) NeedSync(netHeight uint64) (bool, uint64, uint64) {
	//TODO：height，block confirmed Length
	/*
		if netHeight > height+uint64(SyncNumber) {
			return true, height + 1, netHeight
		}
			bc := sy.blockCache.LongestChain()
			ter := bc.Iterator()
			witness := bc.Top().Head.Witness
			num := 0
			for i := 0; i < sy.confirmNumber; i++ {
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
	*/
	return false, 0, 0
}

func (sy *SyncImpl) queryBlockHash(start, end uint64) error {
	hr := message.BlockHashQuery{Start: start, End: end}
	bytes, err := hr.Marshal(nil)
	if err != nil {
		bn.log.D("marshal BlockHashQuery failed. err=%v", err)
		return err
	}
	msg := message.Message{
		Body:    bytes,
		ReqType: int32(BlockHashQuery),
		TTL:     1, //BlockHashQuery req just broadcast to its neibour
		From:    bn.localNode.Addr(),
		Time:    time.Now().UnixNano(),
	}
	bn.log.D("[net] query block hash. start=%v, end=%v, from=%v", start, end, bn.localNode.Addr())
	sy.router.broadcast(msg)
	return nil
}

func (sy *SyncImpl) SyncBlocks(startNumber uint64, endNumber uint64) error {
	var syncNum int
	for endNumber > startNumber+uint64(MaxBlockHashQueryNumber)-1 {
		need := false
		for i := startNumber; i < startNumber+uint64(MaxBlockHashQueryNumber); i++ {
			_, ok := sy.requestMap.LoadOrStore(i, true)
			if !ok {
				need = true
			}
		}
		if need {
			syncNum++
			sy.queryBlockHash(startNumber, startNumber+uint64(MaxBlockHashQueryNumber)-1)
		}
		startNumber += uint64(MaxBlockHashQueryNumber)
		if syncNum%10 == 0 {
			time.Sleep(time.Second)
		}
	}
	if startNumber <= endNumber {
		need := false
		for i := startNumber; i < endNumber; i++ {
			_, ok := sy.requestMap.LoadOrStore(i, true)
			if !ok {
				need = true
			}
		}
		if need {
			sy.router.QueryBlockHash(startNumber, endNumber)
		}
	}
	return nil
}

func (sy *SyncImpl) OnBlockConfirmed(hash, peerID string) error {
	return sy.dc.OnBlockConfirmed(hash, peerID)
}

func (sy *syncImpl) messageLoop() {
	for {
		select {
		case req, ok := <-sy.messageChan:
			if !ok {
				break
			}
			if req.ReqType == network.BlockHashQuery {
				var rh message.BlockHashQuery
				_, err := rh.Unmarshal(req.Body)
				if err != nil {
					sy.log.E("unmarshal BlockHashQuery failed:%v", err)
					break
				}
				go sy.handleHashQuery(rh)
			} else if req.ReqType == network.BlockHashResponse {
				var rh message.BlockHashResponse
				_, err := rh.Unmarshal(req.Body)
				if err != nil {
					sy.log.E("unmarshal BlockHashResponse failed:%v", err)
					break
				}
				go sy.handleHashResp(rh)
			} else if req.ReqType == network.ReqDownloadBlock {
				var rh message.RequestBlock
				err := rh.Decode(req.Body)
				if err != nil {
					break
				}
				go sy.handleBlockQuery(rh)
			}
		case <-sync.exitSignal:
			return
		}
	}

}

func (sy *SyncImpl) handleHashQuery(rh message.BlockHashQuery) {
	if rh.End < rh.Start {
		break
	}
	chain := sy.blockCache.LongestChain()
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
		sy.log.E("marshal BlockHashResponse failed:struct=%v, err=%v", resp, err)
		break
	}
	resMsg := message.Message{
		Time:    time.Now().Unix(),
		From:    req.To,
		To:      req.From,
		ReqType: int32(BlockHashResponse),
		Body:    bytes,
	}
	sy.router.Send(resMsg)
}

func (sy *SyncImpl) handleHashResp(rh *message.BlockHashResponse) {
	sy.log.I("receive block hashes: len=%v", len(rh.BlockHashes))
	for _, blkHash := range rh.BlockHashes {
		if !sy.blockCache.CheckBlock(blkHash.Hash) { // TODO: check hash @ BlockCache and BlockDB
			sy.requestMap.Delete(blkHash.Height)
			sy.dc.OnRecvHash(blkHash.Hash, req.From)
		}
	}
}

func (sy *SyncImpl) retryDownloadLoop() {
	for {
		select {
		case <-time.After(RetryTime):
			sy.requestMap.Range(func(k, v interface{}) bool {
				num, ok := k.(uint64)
				if !ok {
					return false
				}
				if num < sy.blockCache.ConfirmedLength() { // TODO
					sy.requestMap.Delete(num)
				} else {
					sy.queryBlockHash(num, num)
				}
				return true
			})
		case <-sy.exitSignal:
			return
		}
	}
}

func (sy *SyncImpl) handleBlockQuery(rh *message.RequestBlock) {
	chain := sy.blockCache.BlockChain()
	var b []byte
	if rh.BlockNumber < chain.Length() {
		b, err = chain.GetBlockByteByHash(rh.BlockHash)
		if err != nil {
			log.Log.E("Database error: block empty %v", rh.BlockNumber)
			break
		}
	} else {
		block, err := sy.blockCache.FindBlockInCache(rh.BlockHash)
		if err != nil {
			log.Log.E("Block not in cache: %v", rh.BlockNumber)
			break
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
	sy.router.Send(resMsg)
}

type DownloadController interface {
	OnRecvHash(hash, peerid string) error
	OnTimeout() error
	OnBlockConfirmed(hash string, peerID string) error
	DownloadLoop(callback func(hash, peerID string))
	Stop()
}

type DownloadControllerImpl struct {
	hashState  *sync.Map
	peerState  *sync.Map
	peerMap    map[string]*sy.Map
	peerTimer  map[string]*time.Timer
	chDownload chan bool
	exitSignal chan struct{}
}

func NewDownloadController() (*DownloadControllerImpl, error) {
	dc := &DownloadControllerImpl{
		hashState:  new(sy.Map),
		peerState:  new(sy.Map),
		peerMap:    make(map[string]*sy.Map, 0),
		peerTimer:  make(map[string]*time.Timer, 0),
		chDownload: make(chan bool, 100),
		exitSignal: make(chan struct{}),
	}
	return dc, nil
}

func (dc *DownloadControllerImpl) Stop() {
	close(dc.exitSignal)
}

func (dc *DownloadControllerImpl) OnRecvHash(hash string, peerID string) error {
	if _, ok := dc.peerMap[peerID]; !ok {
		hashMap := new(sy.Map)
		dc.peerMap[peerID] = hashMap
	}
	dc.peerMap[peerID].Store(hash, true)
	hState, _ := dc.hashState.LoadOrStore(hash, "Wait")
	pState, _ := dc.peerState.LoadOrStore(peerID, "Free")
	if hState.(string) == "Wait" && pState.(string) == "Free" {
		dc.chDownload <- true
	}
	return nil
}

func (dc *DownloadControllerImpl) OnTimeout(hash string, peerID string) error {
	fmt.Println("OnTimeout", hash, peerID)
	if hState, hok := dc.hashState.Load(hash); hok {
		hs, ok := hState.(string)
		if !ok {
			dc.hashState.Delete(hash)
		} else if hs != "Done" {
			dc.hashState.Store(hash, "Wait")
		}
	}
	if pState, pok := dc.peerState.Load(peerID); pok {
		ps, ok := pState.(string)
		if !ok {
			dc.peerState.Delete(peerID)
		} else if ps == hash {
			dc.peerState.Store(peerID, "Free")
			dc.chDownload <- true
		}
	}
	return nil
}

func (dc *DownloadControllerImpl) OnBlockConfirmed(hash string, peerID string) error {
	fmt.Println("OnRecvBlock", hash, peerID)
	dc.hashState.Store(hash, "Done")
	if pState, pok := dc.peerState.Load(peerID); pok {
		ps, ok := pState.(string)
		if !ok {
			dc.peerState.Delete(peerID)
		} else if ps == hash {
			dc.peerState.Store(peerID, "Free")
			if pTimer, ook := dc.peerTimer[peerID]; ook {
				pTimer.Stop()
			}
			dc.chDownload <- true
		}
	}
	return nil
}

func (dc *DownloadControllerImpl) DownloadLoop(callback func(hash, peerID string)) {
	for {
		select {
		case _, ok := <-dc.chDownload:
			if !ok {
				break
			}
			for peerID, hashMap := range dc.peerMap {
				pState, pok := dc.peerState.Load(peerID)
				if !pok {
					continue
				} else {
					ps, ok := pState.(string)
					if !ok {
						dc.peerState.Delete(peerID)
					} else if ps != "Free" {
						continue
					}
				}
				hashMap.Range(func(k, v interface{}) bool {
					hash, hok := k.(string)
					if !hok {
						return true
					}
					hState, ok := dc.hashState.Load(hash)
					if !ok {
						hashMap.Delete(hash)
						return true
					}
					hste, hsteok := hState.(string)
					if !hsteok {
						dc.hashState.Delete(hash)
						return true
					}
					if hste == "Done" {
						dc.hashState.Delete(hash)
						hashMap.Delete(hash)
						return true
					}
					if hste == "Wait" {
						dc.peerState.Store(peerID, hash)
						dc.hashState.Store(hash, peerID)
						callback(hash, peerID)
						dc.peerTimer[peerID] = time.AfterFunc(5*time.Second, func() {
							dc.OnTimeout(hash, peerID)
						})
						return false
					}
					return true
				})
			}
		case <-dc.exitSignal:
			//fmt.Println("Close DownloadLoop")
			return
		}
	}
}
