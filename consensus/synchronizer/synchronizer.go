package synchronizer

import (
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	block "github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
)

var (
	SyncNumber              int64 = 2
	MaxBlockHashQueryNumber int64 = 10
	RetryTime                     = 5 * time.Second
	MaxAcceptableLength     int64 = 100
	ConfirmNumber           int   = 7
)

type Synchronizer interface {
	Start() error
	Stop()
	NeedSync(maxHeight int64) (bool, int64, int64)
	SyncBlocks(startNumber int64, endNumber int64) error
	OnBlockConfirmed(hash string, peerID p2p.PeerID) error
}

type SyncImpl struct {
	p2pService p2p.Service
	blockCache blockcache.BlockCache
	lastHead   *blockcache.BlockCacheNode
	glb        global.BaseVariable
	dc         DownloadController
	reqMap     *sync.Map

	messageChan chan p2p.IncomingMessage
	exitSignal  chan struct{}
}

func NewSynchronizer(glb global.BaseVariable, blkcache blockcache.BlockCache, p2pserv p2p.Service) (*SyncImpl, error) {
	sy := &SyncImpl{
		p2pService: p2pserv,
		blockCache: blkcache,
		glb:        glb,
		reqMap:     new(sync.Map),
		lastHead:   nil,
	}
	var err error
	sy.dc, err = NewDownloadController()
	if err != nil {
		return nil, err
	}

	sy.messageChan = sy.p2pService.Register("sync message",
		p2p.SyncBlockRequest,
		p2p.SyncBlockHashRequest,
		p2p.SyncBlockHashResponse,
	)

	sy.exitSignal = make(chan struct{})

	return sy, nil
}

func (sy *SyncImpl) reqDownloadBlock(hash string, peerID p2p.PeerID) {
	blkReq := &message.RequestBlock{
		BlockHash: []byte(hash),
	}
	bytes, err := blkReq.Encode()
	if err != nil {
		return
	}
	sy.p2pService.SendToPeer(peerID, bytes, p2p.SyncBlockRequest, p2p.UrgentMessage)
}

func (sy *SyncImpl) Start() error {
	go sy.dc.DownloadLoop(sy.reqDownloadBlock)
	go sy.messageLoop()
	go sy.retryDownloadLoop()
	return nil
}

func (sy *SyncImpl) Stop() {
	sy.dc.Stop()
	close(sy.exitSignal)
}

func (sy *SyncImpl) NeedSync(netHeight int64) (bool, int64, int64) {
	height := sy.glb.BlockChain().Length() - 1
	if netHeight > height+SyncNumber {
		return true, height + 1, netHeight
	}
	bcn := sy.blockCache.Head()
	num := 0
	if bcn != sy.lastHead {
		sy.lastHead = sy.blockCache.Head()
		witness := bcn.Block.Head.Witness
		for i := 0; i < ConfirmNumber; i++ {
			bcn = bcn.Parent
			if bcn == nil {
				break
			}
			if witness == bcn.Block.Head.Witness {
				num++
			}
		}
	}
	if num > 0 {
		return true, height + 1, netHeight
	}
	return false, 0, 0
}

func (sy *SyncImpl) queryBlockHash(start, end int64) error {
	hr := message.BlockHashQuery{ReqType: 0, Start: start, End: end}
	bytes, err := hr.Encode()
	if err != nil {
		ilog.Debug("marshal BlockHashQuery failed. err=%v", err)
		return err
	}
	ilog.Debug("[net] query block hash. start=%v, end=%v", start, end)
	sy.p2pService.Broadcast(bytes, p2p.SyncBlockHashRequest, p2p.UrgentMessage)
	return nil
}

func (sy *SyncImpl) SyncBlocks(startNumber int64, endNumber int64) error {
	var syncNum int
	for endNumber > startNumber+MaxBlockHashQueryNumber-1 {
		need := false
		for i := startNumber; i < startNumber+MaxBlockHashQueryNumber; i++ {
			_, ok := sy.reqMap.LoadOrStore(i, true)
			if !ok {
				need = true
			}
		}
		if need {
			syncNum++
			sy.queryBlockHash(startNumber, startNumber+MaxBlockHashQueryNumber-1)
		}
		startNumber += MaxBlockHashQueryNumber
		if syncNum%10 == 0 {
			time.Sleep(time.Second)
		}
	}
	if startNumber <= endNumber {
		need := false
		for i := startNumber; i < endNumber; i++ {
			_, ok := sy.reqMap.LoadOrStore(i, true)
			if !ok {
				need = true
			}
		}
		if need {
			sy.queryBlockHash(startNumber, endNumber)
		}
	}
	return nil
}

func (sy *SyncImpl) OnBlockConfirmed(hash string, peerID p2p.PeerID) error {
	return sy.dc.OnBlockConfirmed(hash, peerID)
}

func (sy *SyncImpl) messageLoop() {
	for {
		select {
		case req, ok := <-sy.messageChan:
			if !ok {
				return
			}
			if req.Type() == p2p.SyncBlockHashRequest {
				var rh message.BlockHashQuery
				err := rh.Decode(req.Data())
				if err != nil {
					ilog.Debug("unmarshal BlockHashQuery failed:%v", err)
					break
				}
				go sy.handleHashQuery(&rh, req.From())
			} else if req.Type() == p2p.SyncBlockHashResponse {
				var rh message.BlockHashResponse
				err := rh.Decode(req.Data())
				if err != nil {
					ilog.Debug("unmarshal BlockHashResponse failed:%v", err)
					break
				}
				go sy.handleHashResp(&rh, req.From())
			} else if req.Type() == p2p.SyncBlockRequest {
				var rh message.RequestBlock
				err := rh.Decode(req.Data())
				if err != nil {
					break
				}
				go sy.handleBlockQuery(&rh, req.From())
			}
		case <-sy.exitSignal:
			return
		}
	}

}

func (sy *SyncImpl) handleHashQuery(rh *message.BlockHashQuery, peerID p2p.PeerID) {
	if rh.End < rh.Start {
		return
	}
	resp := &message.BlockHashResponse{
		BlockHashes: make([]*message.BlockHash, 0, rh.End-rh.Start+1),
	}
	if rh.ReqType == 0 {
		node := sy.blockCache.Head()
		var i int64
		for i = rh.End; i >= rh.Start; i-- {
			var hash []byte
			var err error
			if i < sy.blockCache.LinkedRoot().Number {
				hash, err = sy.glb.BlockChain().GetHashByNumber(i)
				if err != nil {
					continue
				}
			} else {
				if i > node.Number {
					continue
				}
				for node != nil && i < node.Number {
					node = node.Parent
				}
				if node == nil || i != node.Number {
					continue
				}
				hash = node.Block.HeadHash()
			}
			blkHash := message.BlockHash{
				Height: i,
				Hash:   hash,
			}
			resp.BlockHashes = append(resp.BlockHashes, &blkHash)
		}
	} else {
		var blk *block.Block
		var err error
		for _, num := range rh.Nums {
			var hash []byte
			blk, err = sy.blockCache.GetBlockByNumber(num)
			if err == nil {
				hash = blk.HeadHash()
			} else {
				hash, err = sy.glb.BlockChain().GetHashByNumber(num)
				if err != nil {
					continue
				}
			}
			blkHash := message.BlockHash{
				Height: num,
				Hash:   hash,
			}
			resp.BlockHashes = append(resp.BlockHashes, &blkHash)
		}
	}
	if len(resp.BlockHashes) == 0 {
		return
	}
	bytes, err := resp.Encode()
	if err != nil {
		ilog.Error("marshal BlockHashResponse failed:struct=%v, err=%v", resp, err)
		return
	}
	sy.p2pService.SendToPeer(peerID, bytes, p2p.SyncBlockHashResponse, p2p.NormalMessage)
}

func (sy *SyncImpl) handleHashResp(rh *message.BlockHashResponse, peerID p2p.PeerID) {
	ilog.Info("receive block hashes: len=%v", len(rh.BlockHashes))
	for _, blkHash := range rh.BlockHashes {
		if _, err := sy.blockCache.Find(blkHash.Hash); err == nil { // TODO: check hash @ BlockCache and BlockDB
			sy.reqMap.Delete(blkHash.Height)
			sy.dc.OnRecvHash(string(blkHash.Hash), peerID)
		}
	}
}

func (sy *SyncImpl) retryDownloadLoop() {
	for {
		select {
		case <-time.After(RetryTime):
			sy.reqMap.Range(func(k, v interface{}) bool {
				num, ok := k.(int64)
				if !ok {
					return false
				}
				if num <= sy.blockCache.LinkedRoot().Number { // TODO
					sy.reqMap.Delete(num)
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

func (sy *SyncImpl) handleBlockQuery(rh *message.RequestBlock, peerID p2p.PeerID) {
	var b []byte
	var err error
	if int64(rh.BlockNumber) < sy.blockCache.LinkedRoot().Number {
		b, err = sy.glb.BlockChain().GetBlockByteByHash(rh.BlockHash)
		if err != nil {
			ilog.Error("Database error: block empty %v", rh.BlockNumber)
			return
		}
	} else {
		node, err := sy.blockCache.Find(rh.BlockHash)
		if err != nil {
			ilog.Error("Block not in cache: %v", rh.BlockNumber)
			return
		}
		b, err = node.Block.Encode()
		if err != nil {
			ilog.Error("Fail to encode block: %v", rh.BlockNumber)
			return
		}
	}
	sy.p2pService.SendToPeer(peerID, b, p2p.SyncBlockResponse, p2p.NormalMessage)
}

type DownloadController interface {
	OnRecvHash(hash string, peerID p2p.PeerID) error
	OnTimeout(hash string, peerID p2p.PeerID) error
	OnBlockConfirmed(hash string, peerID p2p.PeerID) error
	DownloadLoop(callback func(hash string, peerID p2p.PeerID))
	Stop()
}

type DownloadControllerImpl struct {
	hashState  *sync.Map
	peerState  *sync.Map
	peerMap    map[p2p.PeerID]*sync.Map
	peerTimer  map[p2p.PeerID]*time.Timer
	chDownload chan bool
	exitSignal chan struct{}
}

func NewDownloadController() (*DownloadControllerImpl, error) {
	dc := &DownloadControllerImpl{
		hashState:  new(sync.Map),
		peerState:  new(sync.Map),
		peerMap:    make(map[p2p.PeerID]*sync.Map, 0),
		peerTimer:  make(map[p2p.PeerID]*time.Timer, 0),
		chDownload: make(chan bool, 100),
		exitSignal: make(chan struct{}),
	}
	return dc, nil
}

func (dc *DownloadControllerImpl) Stop() {
	close(dc.exitSignal)
}

func (dc *DownloadControllerImpl) OnRecvHash(hash string, peerID p2p.PeerID) error {
	if _, ok := dc.peerMap[peerID]; !ok {
		hashMap := new(sync.Map)
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

func (dc *DownloadControllerImpl) OnTimeout(hash string, peerID p2p.PeerID) error {
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

func (dc *DownloadControllerImpl) OnBlockConfirmed(hash string, peerID p2p.PeerID) error {
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

func (dc *DownloadControllerImpl) DownloadLoop(callback func(hash string, peerID p2p.PeerID)) {
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
			return
		}
	}
}
