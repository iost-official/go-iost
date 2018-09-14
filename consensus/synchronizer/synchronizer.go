package synchronizer

import (
	"sort"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
)

var (
	// TODO: configurable
	confirmNumber = 3
	// SyncNumber    int64 = int64(ConfirmNumber) * 2 / 3
	syncNumber int64 = 11

	maxBlockHashQueryNumber int64 = 1000
	retryTime                     = 5 * time.Second
	syncBlockTimeout              = time.Second
	syncHeightTime                = 3 * time.Second
	heightTimeout           int64 = 5 * 3
)

type Synchronizer interface {
	Start() error
	Stop()
	CheckSync() bool
	CheckGenBlock(hash []byte) bool
	SyncBlocks(startNumber int64, endNumber int64) error
	OnBlockConfirmed(hash string)
	OnRecvBlock(hash string, peerID p2p.PeerID)
	CheckSyncProcess()
}

type SyncImpl struct {
	p2pService   p2p.Service
	blockCache   blockcache.BlockCache
	lastHead     *blockcache.BlockCacheNode
	basevariable global.BaseVariable
	dc           DownloadController
	reqMap       *sync.Map
	heightMap    *sync.Map
	syncEnd      int64

	messageChan    chan p2p.IncomingMessage
	syncHeightChan chan p2p.IncomingMessage
	exitSignal     chan struct{}
}

func NewSynchronizer(basevariable global.BaseVariable, blkcache blockcache.BlockCache, p2pserv p2p.Service) (*SyncImpl, error) {
	sy := &SyncImpl{
		p2pService:   p2pserv,
		blockCache:   blkcache,
		basevariable: basevariable,
		reqMap:       new(sync.Map),
		heightMap:    new(sync.Map),
		lastHead:     nil,
		syncEnd:      0,
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

	sy.syncHeightChan = sy.p2pService.Register("sync height", p2p.SyncHeight)
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
	go sy.syncHeightLoop()
	go sy.messageLoop()
	go sy.retryDownloadLoop()
	go sy.Initializer()
	return nil
}

func (sy *SyncImpl) Stop() {
	sy.dc.Stop()
	close(sy.exitSignal)
}

func (sy *SyncImpl) Initializer() {
	if sy.basevariable.Mode() != global.ModeInit {
		return
	}
	for {
		select {
		case <-time.After(retryTime):
			if sy.basevariable.BlockChain().Length() == 0 {
				sy.SyncBlocks(0, 0)
				continue
			} else {
				sy.basevariable.SetMode(global.ModeNormal)
				sy.CheckSync()
				return
			}
		case <-sy.exitSignal:
			return
		}

	}

}

func (sy *SyncImpl) syncHeightLoop() {
	for {
		select {
		case <-time.After(syncHeightTime):
			num := sy.blockCache.Head().Number
			sh := &message.SyncHeight{Height: num, Time: time.Now().Unix()}
			bytes, err := proto.Marshal(sh)
			if err != nil {
				ilog.Errorf("marshal syncheight failed. err=%v", err)
				continue
			}
			sy.p2pService.Broadcast(bytes, p2p.SyncHeight, p2p.NormalMessage)
		case req := <-sy.syncHeightChan:
			var sh message.SyncHeight
			err := proto.UnmarshalMerge(req.Data(), &sh)
			if err != nil {
				ilog.Errorf("unmarshal syncheight failed. err=%v", err)
				continue
			}
			// ilog.Debugf("sync height from: %s, height: %v, time:%v", req.From().Pretty(), sh.Height, sh.Time)
			sy.heightMap.Store(req.From(), sh)
		case <-sy.exitSignal:
			return
		}
	}

}

func (sy *SyncImpl) CheckSync() bool {
	if sy.basevariable.Mode() != global.ModeNormal {
		return false
	}
	height := sy.basevariable.BlockChain().Length() - 1
	heights := make([]int64, 0, 0)
	heights = append(heights, sy.blockCache.Head().Number)
	now := time.Now().Unix()
	sy.heightMap.Range(func(k, v interface{}) bool {
		sh, ok := v.(message.SyncHeight)
		if !ok || sh.Time+heightTimeout < now {
			sy.heightMap.Delete(k)
			return true
		}
		heights = append(heights, 0)
		r := len(heights) - 1
		for 0 < r && heights[r-1] > sh.Height {
			heights[r] = heights[r-1]
			r--
		}
		heights[r] = sh.Height
		return true
	})
	// ilog.Debugf("check sync heights: %v", heights)
	netHeight := heights[len(heights)/2]
	if netHeight > height+syncNumber {
		sy.basevariable.SetMode(global.ModeSync)
		sy.dc.Reset()
		go sy.SyncBlocks(height+1, netHeight)
		return true
	}
	return false
}

func (sy *SyncImpl) CheckGenBlock(hash []byte) bool {
	if sy.basevariable.Mode() != global.ModeNormal {
		return false
	}
	bcn, err := sy.blockCache.Find(hash)
	if err != nil {
		ilog.Errorf("bcn not found on chekc gen block, err:%s", err)
		return false
	}
	height := sy.basevariable.BlockChain().Length() - 1
	num := 0
	if bcn != sy.lastHead {
		sy.lastHead = sy.blockCache.Head()
		witness := bcn.Block.Head.Witness
		for i := 0; i < confirmNumber; i++ {
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
		go sy.SyncBlocks(height+1, sy.blockCache.Head().Number)
		return true
	}
	return false
}

func (sy *SyncImpl) queryBlockHash(hr *message.BlockHashQuery) {
	bytes, err := hr.Encode()
	if err != nil {
		ilog.Errorf("marshal blockhashquery failed. err=%v", err)
		return
	}
	ilog.Infof("[sync] request block hash. reqtype=%v, start=%v, end=%v, nums size=%v", hr.ReqType, hr.Start, hr.End, len(hr.Nums))
	sy.p2pService.Broadcast(bytes, p2p.SyncBlockHashRequest, p2p.UrgentMessage)
}

func (sy *SyncImpl) SyncBlocks(startNumber int64, endNumber int64) error {
	sy.syncEnd = endNumber
	for endNumber > startNumber+maxBlockHashQueryNumber-1 {
		for sy.blockCache.Head().Number+3 < startNumber {
			time.Sleep(500 * time.Millisecond)
		}
		for i := startNumber; i < startNumber+maxBlockHashQueryNumber; i++ {
			sy.reqMap.Store(i, true)
		}
		sy.queryBlockHash(&message.BlockHashQuery{ReqType: 0, Start: startNumber, End: startNumber + maxBlockHashQueryNumber - 1, Nums: nil})
		startNumber += maxBlockHashQueryNumber
	}
	if startNumber <= endNumber {
		for i := startNumber; i <= endNumber; i++ {
			sy.reqMap.Store(i, true)
		}
		sy.queryBlockHash(&message.BlockHashQuery{ReqType: 0, Start: startNumber, End: endNumber, Nums: nil})
	}
	return nil
}

func (sy *SyncImpl) CheckSyncProcess() {
	if sy.syncEnd <= sy.blockCache.Head().Number {
		sy.basevariable.SetMode(global.ModeNormal)
		sy.dc.Reset()
	} else {
		ilog.Infof("check sync process: now %v, end %v", sy.blockCache.Head().Number, sy.syncEnd)
	}
}

func (sy *SyncImpl) OnBlockConfirmed(hash string) {
	sy.dc.MissionComplete(hash)
}

func (sy *SyncImpl) OnRecvBlock(hash string, peerID p2p.PeerID) {
	sy.dc.FreePeer(hash, peerID)
}

func (sy *SyncImpl) messageLoop() {
	for {
		select {
		case req := <-sy.messageChan:
			if req.Type() == p2p.SyncBlockHashRequest {
				var rh message.BlockHashQuery
				err := rh.Decode(req.Data())
				if err != nil {
					ilog.Errorf("unmarshal BlockHashQuery failed:%v", err)
					break
				}
				go sy.handleHashQuery(&rh, req.From())
			} else if req.Type() == p2p.SyncBlockHashResponse {
				var rh message.BlockHashResponse
				err := rh.Decode(req.Data())
				if err != nil {
					ilog.Errorf("unmarshal BlockHashResponse failed:%v", err)
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
	if rh.End < rh.Start || rh.Start < 0 {
		return
	}
	resp := &message.BlockHashResponse{
		BlockHashes: make([]*message.BlockHash, 0, rh.End-rh.Start+1),
	}
	if rh.ReqType == 0 {
		node := sy.blockCache.Head()

		var end = rh.End
		if node != nil && end > node.Number {
			end = node.Number
		}

		for i := end; i >= rh.Start; i-- {
			var hash []byte
			var err error

			for node != nil && i < node.Number {
				node = node.Parent
			}

			if node != nil {
				hash = node.Block.HeadHash()
			} else {
				hash, err = sy.basevariable.BlockChain().GetHashByNumber(i)
				if err != nil {
					ilog.Errorf("get hash by number from db failed. err=%v, number=%v", err, i)
					continue
				}
			}

			blkHash := message.BlockHash{
				Height: i,
				Hash:   hash,
			}
			resp.BlockHashes = append(resp.BlockHashes, &blkHash)
		}
		for i, j := 0, len(resp.BlockHashes)-1; i < j; i, j = i+1, j-1 {
			resp.BlockHashes[i], resp.BlockHashes[j] = resp.BlockHashes[j], resp.BlockHashes[i]
		}
	}
	if rh.ReqType == 1 {
		var blk *block.Block
		var err error
		for _, num := range rh.Nums {
			var hash []byte
			blk, err = sy.blockCache.GetBlockByNumber(num)
			if err == nil {
				hash = blk.HeadHash()
			} else {
				hash, err = sy.basevariable.BlockChain().GetHashByNumber(num)
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
		ilog.Errorf("marshal BlockHashResponse failed:struct=%v, err=%v", resp, err)
		return
	}
	sy.p2pService.SendToPeer(peerID, bytes, p2p.SyncBlockHashResponse, p2p.NormalMessage)
}

func (sy *SyncImpl) handleHashResp(rh *message.BlockHashResponse, peerID p2p.PeerID) {
	ilog.Infof("receive block hashes: len=%v", len(rh.BlockHashes))
	for _, blkHash := range rh.BlockHashes {
		sy.reqMap.Delete(blkHash.Height)
		if blkHash.Height < sy.basevariable.BlockChain().Length() {
			continue
		}
		if _, err := sy.blockCache.Find(blkHash.Hash); err != nil {
			sy.dc.OnRecvHash(string(blkHash.Hash), peerID)
		}
	}
}

func (sy *SyncImpl) retryDownloadLoop() {
	for {
		select {
		case <-time.After(retryTime):
			hq := &message.BlockHashQuery{ReqType: 1, Start: 0, End: 0, Nums: make([]int64, 0)}
			sy.reqMap.Range(func(k, v interface{}) bool {
				num, ok := k.(int64)
				if !ok {
					sy.reqMap.Delete(k)
					return true
				}
				hq.Nums = append(hq.Nums, num)
				return true
			})
			if len(hq.Nums) > 0 {
				// ilog.Debug("retry download ", hr.Nums)
				sort.Slice(hq.Nums, func(i int, j int) bool {
					return hq.Nums[i] < hq.Nums[j]
				})
				sy.queryBlockHash(hq)
			}
		case <-sy.exitSignal:
			return
		}
	}
}

func (sy *SyncImpl) handleBlockQuery(rh *message.RequestBlock, peerID p2p.PeerID) {
	var b []byte
	var err error
	node, err := sy.blockCache.Find(rh.BlockHash)
	if err == nil {
		b, err = node.Block.Encode()
		if err != nil {
			ilog.Errorf("Fail to encode block: %v, err=%v", rh.BlockNumber, err)
			return
		}
		sy.p2pService.SendToPeer(peerID, b, p2p.SyncBlockResponse, p2p.NormalMessage)
		return
	}
	b, err = sy.basevariable.BlockChain().GetBlockByteByHash(rh.BlockHash)
	if err != nil {
		ilog.Errorf("handle block query failed to get block.")
		return
	}
	sy.p2pService.SendToPeer(peerID, b, p2p.SyncBlockResponse, p2p.NormalMessage)
}

type DownloadController interface {
	OnRecvHash(hash string, peerID p2p.PeerID)
	OnTimeout(hash string, peerID p2p.PeerID)
	MissionComplete(hash string)
	FreePeer(hash string, peerID p2p.PeerID)
	DownloadLoop(callback func(hash string, peerID p2p.PeerID))
	Reset()
	Stop()
}

const (
	Done string = "Done"
	Wait string = "Wait"
)

const (
	peerConNum        = 10
	Free       string = "Free"
)

const (
	Head string = "Head"
	Tail string = "Tail"
)

type hashListNode struct {
	val  string
	prev *hashListNode
	next *hashListNode
}

type DownloadControllerImpl struct {
	hashState      *sync.Map
	peerState      *sync.Map
	peerStateMutex *sync.Map
	peerMap        *sync.Map
	peerMapMutex   *sync.Map
	chDownload     chan struct{}
	exitSignal     chan struct{}
}

func NewDownloadController() (*DownloadControllerImpl, error) {
	dc := &DownloadControllerImpl{
		hashState:      new(sync.Map), // map[string]string
		peerState:      new(sync.Map), // map[PeerID](map[string]bool)
		peerStateMutex: new(sync.Map), // map[PeerID](metux)
		peerMap:        new(sync.Map), // map[PeerID](map[string]bool)
		peerMapMutex:   new(sync.Map), // map[PeerID](map[string]bool)
		chDownload:     make(chan struct{}, 2),
		exitSignal:     make(chan struct{}),
	}
	return dc, nil
}

func (dc *DownloadControllerImpl) Reset() {
	dc.hashState = new(sync.Map)
	dc.peerState = new(sync.Map)
	dc.peerStateMutex = new(sync.Map)
	dc.peerMap = new(sync.Map)
	dc.peerMapMutex = new(sync.Map)
}

func (dc *DownloadControllerImpl) Stop() {
	close(dc.exitSignal)
}

func (dc *DownloadControllerImpl) OnRecvHash(hash string, peerID p2p.PeerID) {
	// ilog.Debugf("peer: %s, hash: %s", peerID, hash)
	hStateIF, _ := dc.hashState.LoadOrStore(hash, Wait)

	_, ok := dc.peerState.Load(peerID)
	if !ok {
		pState := make(map[string]*time.Timer)
		pmMutex, _ := dc.peerMapMutex.LoadOrStore(peerID, new(sync.Mutex))
		hm, ok := dc.peerMap.LoadOrStore(peerID, new(sync.Map))
		if !ok {
			pmMutex.(*sync.Mutex).Lock()
			head := &hashListNode{val: Head, prev: nil, next: nil}
			tail := &hashListNode{val: Tail, prev: nil, next: nil}
			head.next = tail
			tail.prev = head
			hashMap, _ := hm.(*sync.Map)
			hashMap.Store(head.val, head)
			hashMap.Store(tail.val, tail)
			pmMutex.(*sync.Mutex).Unlock()
		}
		dc.peerStateMutex.LoadOrStore(peerID, new(sync.Mutex))
		dc.peerState.LoadOrStore(peerID, pState)
	}
	hmIF, _ := dc.peerMap.Load(peerID)
	hashMap, _ := hmIF.(*sync.Map)
	if _, ok = hashMap.Load(hash); !ok {
		pmMutex, _ := dc.peerMapMutex.Load(peerID)
		tailNode, _ := hashMap.Load(Tail)
		tail, _ := tailNode.(*hashListNode)
		pmMutex.(*sync.Mutex).Lock()
		node := &hashListNode{val: hash, prev: tail.prev, next: tail}
		node.prev.next = node
		node.next.prev = node
		hashMap.Store(node.val, node)
		pmMutex.(*sync.Mutex).Unlock()
	}
	if hState, ok := hStateIF.(string); ok && hState == Wait {
		select {
		case dc.chDownload <- struct{}{}:
		default:
		}
	}
}

func (dc *DownloadControllerImpl) OnTimeout(hash string, peerID p2p.PeerID) {
	ilog.Debugf("sync timout, hash=%v, peerID=%s", []byte(hash), peerID.Pretty())
	if hStateIF, ok := dc.hashState.Load(hash); ok {
		hState, ok := hStateIF.(string)
		if !ok {
			dc.hashState.Delete(hash)
		} else if hState != Done {
			dc.hashState.Store(hash, Wait)
		}
	}
	if pStateIF, ok := dc.peerState.Load(peerID); ok {
		psMutex, _ := dc.peerStateMutex.Load(peerID)
		psMutex.(*sync.Mutex).Lock()
		pState, ok := pStateIF.(map[string]bool)
		if !ok {
			dc.peerState.Delete(peerID)
		} else {
			if _, ok = pState[hash]; ok {
				delete(pState, hash)
				select {
				case dc.chDownload <- struct{}{}:
				default:
				}
			}
		}
		psMutex.(*sync.Mutex).Unlock()
	}
}

func (dc *DownloadControllerImpl) MissionComplete(hash string) {
	dc.hashState.Store(hash, Done)
}

func (dc *DownloadControllerImpl) FreePeer(hash string, peerID p2p.PeerID) {
	if pStateIF, ok := dc.peerState.Load(peerID); ok {
		psMutex, _ := dc.peerStateMutex.Load(peerID)
		psMutex.(*sync.Mutex).Lock()
		pState, ok := pStateIF.(map[string]*time.Timer)
		if !ok {
			dc.peerState.Delete(peerID)
		} else {
			if timer, ok := pState[hash]; ok {
				timer.Stop()
				delete(pState, hash)
				select {
				case dc.chDownload <- struct{}{}:
				default:
				}
			}
		}
		psMutex.(*sync.Mutex).Unlock()
	}

}

func (dc *DownloadControllerImpl) DownloadLoop(callback func(hash string, peerID p2p.PeerID)) {
	for {
		select {
		case <-dc.chDownload:
			ilog.Debugf("Download Begin")
			dc.peerState.Range(func(k, v interface{}) bool {
				peerID := k.(p2p.PeerID)
				ps, ok := v.(map[string]*time.Timer)
				if !ok {
					dc.peerState.Delete(peerID)
					return true
				}
				psMutex, _ := dc.peerStateMutex.Load(peerID)
				psMutex.(*sync.Mutex).Lock()
				ilog.Debugf("peerNum: %v", len(ps))
				psLen := len(ps)
				psMutex.(*sync.Mutex).Unlock()
				if psLen >= peerConNum {
					return true
				}

				pmMutex, _ := dc.peerMapMutex.Load(peerID)
				hm, _ := dc.peerMap.Load(peerID)
				hashMap, _ := hm.(*sync.Map)

				pmMutex.(*sync.Mutex).Lock()
				headNode, _ := hashMap.Load(Head)
				node := headNode.(*hashListNode).next
				pmMutex.(*sync.Mutex).Unlock()
				for {
					if node.val == Tail {
						break
					}
					hash := node.val
					var hState string
					hStateIF, ok := dc.hashState.Load(hash)
					if ok {
						hState, ok = hStateIF.(string)
					}
					if !ok || hState == Done {
						dc.hashState.Delete(hash)
						pmMutex.(*sync.Mutex).Lock()
						hashMap.Delete(hash)
						node.prev.next = node.next
						node.next.prev = node.prev
						pmMutex.(*sync.Mutex).Unlock()
					} else if hState == Wait {
						dc.hashState.Store(hash, peerID.Pretty())
						callback(hash, peerID)
						psMutex.(*sync.Mutex).Lock()
						ps[hash] = time.AfterFunc(syncBlockTimeout, func() {
							dc.OnTimeout(hash, peerID)
						})
						psLen := len(ps)
						psMutex.(*sync.Mutex).Unlock()
						if psLen >= peerConNum {
							break
						}
					}
					pmMutex.(*sync.Mutex).Lock()
					node = node.next
					pmMutex.(*sync.Mutex).Unlock()
				}
				return true
			})
			ilog.Debugf("Download End")
		case <-dc.exitSignal:
			return
		}
	}
}
