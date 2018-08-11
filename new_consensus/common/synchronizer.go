package consensus_common

import (
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
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
	OnBlockConfirmed(hash, peerID p2p.PeerID) error
}

type SyncImpl struct {
	p2pService p2p.Service
	blockCache blockcache.BlockCache
	glb        global.Global
	dc         DownloadController
	reqMap     *sync.Map

	messageChan chan p2p.IncomingMessage
	exitSignal  chan struct{}

	log *log.Logger
}

func NewSynchronizer(glb global.Global, blkcache blockcache.BlockCache, p2pserv p2p.Service) (*SyncImpl, error) {
	sy := &SyncImpl{
		p2pService: p2pserv,
		blockCache: blkcache,
		glb:        glb,
		reqMap:     new(sync.Map),
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

	sy.log, err = log.NewLogger("synchronizer.log")
	if err != nil {
		return nil, err
	}

	sy.log.NeedPrint = false
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
	/*
		reqMsg := message.Message{
			Time:    time.Now().Unix(),
			To:      peerID,
			ReqType: int32(p2p.ReqDownloadBlock),
			Body:    blkReq.Encode(),
		}
		sy.router.Send(reqMsg)
	*/
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
	bytes, err := hr.Encode()
	if err != nil {
		sy.log.D("marshal BlockHashQuery failed. err=%v", err)
		return err
	}
	sy.log.D("[net] query block hash. start=%v, end=%v", start, end)
	sy.p2pService.Broadcast(bytes, p2p.SyncBlockHashRequest, p2p.UrgentMessage)
	/*
		msg := message.Message{
			Body:    bytes,
			ReqType: int32(p2p.BlockHashQuery),
			TTL:     1, //BlockHashQuery req just broadcast to its neibour
			Time:    time.Now().UnixNano(),
		}
		sy.router.Broadcast(msg)
	*/
	return nil
}

func (sy *SyncImpl) SyncBlocks(startNumber uint64, endNumber uint64) error {
	var syncNum int
	for endNumber > startNumber+uint64(MaxBlockHashQueryNumber)-1 {
		need := false
		for i := startNumber; i < startNumber+uint64(MaxBlockHashQueryNumber); i++ {
			_, ok := sy.reqMap.LoadOrStore(i, true)
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
				break
			}
			if req.Type() == p2p.SyncBlockHashRequest {
				var rh message.BlockHashQuery
				err := rh.Decode(req.Data())
				if err != nil {
					sy.log.E("unmarshal BlockHashQuery failed:%v", err)
					break
				}
				go sy.handleHashQuery(&rh, req.From())
			} else if req.Type() == p2p.SyncBlockHashResponse {
				var rh message.BlockHashResponse
				err := rh.Decode(req.Data())
				if err != nil {
					sy.log.E("unmarshal BlockHashResponse failed:%v", err)
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
	node := sy.blockCache.Head()
	for i := rh.End; i >= rh.Start; i-- {
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
			hash, err = node.Block.HeadHash()
			if err != nil {
				continue
			}
		}
		blkHash := message.BlockHash{
			Height: i,
			Hash:   hash,
		}
		resp.BlockHashes = append(resp.BlockHashes, &blkHash)
	}
	if len(resp.BlockHashes) == 0 {
		return
	}
	bytes, err := resp.Encode()
	if err != nil {
		sy.log.E("marshal BlockHashResponse failed:struct=%v, err=%v", resp, err)
		return
	}
	sy.p2pService.SendToPeer(peerID, bytes, p2p.SyncBlockHashResponse, p2p.NormalMessage)
	/*
		resMsg := message.Message{
			Time:    time.Now().Unix(),
			To:      peerID,
			ReqType: int32(p2p.BlockHashResponse),
			Body:    bytes,
		}
		sy.router.Send(resMsg)
	*/
}

func (sy *SyncImpl) handleHashResp(rh *message.BlockHashResponse, peerID p2p.PeerID) {
	sy.log.I("receive block hashes: len=%v", len(rh.BlockHashes))
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
				num, ok := k.(uint64)
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
	if rh.BlockNumber < sy.blockCache.LinkedRoot().Number {
		b, err = sy.glb.BlockChain().GetBlockByteByHash(rh.BlockHash)
		if err != nil {
			log.Log.E("Database error: block empty %v", rh.BlockNumber)
			return
		}
	} else {
		node, err := sy.blockCache.Find(rh.BlockHash)
		if err != nil {
			log.Log.E("Block not in cache: %v", rh.BlockNumber)
			return
		}
		b, err = node.Block.Encode()
		if err != nil {
			log.Log.E("Fail to encode block: %v", rh.BlockNumber)
			return
		}
	}
	sy.p2pService.SendToPeer(peerID, b, p2p.SyncBlockResponse, p2p.NormalMessage)
	/*
		resMsg := message.Message{
			Time:    time.Now().Unix(),
			To:      peerID,
			ReqType: int32(p2p.ReqSyncBlock),
			Body:    b,
		}
		sy.router.Send(resMsg)
	*/
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
