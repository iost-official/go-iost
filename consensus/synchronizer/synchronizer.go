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

type callbackfunc = func(hash string, peerID p2p.PeerID)

// Synchronizer defines the functions of synchronizer module
type Synchronizer interface {
	Start() error
	Stop()
	CheckSync() bool
	CheckGenBlock(hash []byte) bool
	OnBlockConfirmed(hash string)
	OnRecvBlock(hash string, peerID p2p.PeerID)
	CheckSyncProcess()
}

//SyncImpl is the implementation of Synchronizer.
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

// NewSynchronizer returns a SyncImpl instance.
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
	sy.dc, err = NewDownloadController(sy.reqDownloadBlock)
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

// Start starts the synchronizer module.
func (sy *SyncImpl) Start() error {
	go sy.dc.Start()
	go sy.syncHeightLoop()
	go sy.messageLoop()
	go sy.retryDownloadLoop()
	go sy.initializer()
	return nil
}

// Stop stops the synchronizer module.
func (sy *SyncImpl) Stop() {
	sy.dc.Stop()
	close(sy.exitSignal)
}

func (sy *SyncImpl) initializer() {
	if sy.basevariable.Mode() != global.ModeInit {
		return
	}
	for {
		select {
		case <-time.After(retryTime):
			if sy.basevariable.BlockChain().Length() == 0 {
				sy.syncBlocks(0, 0)
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
			sy.p2pService.Broadcast(bytes, p2p.SyncHeight, p2p.UrgentMessage)
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

// CheckSync checks if we need to sync.
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
		go sy.syncBlocks(height+1, netHeight)
		return true
	}
	return false
}

// CheckGenBlock checks if we need to sync after gen a block.
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
		go sy.syncBlocks(height+1, sy.blockCache.Head().Number)
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

func (sy *SyncImpl) syncBlocks(startNumber int64, endNumber int64) error {
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

// CheckSyncProcess checks if the end of sync.
func (sy *SyncImpl) CheckSyncProcess() {
	if sy.syncEnd <= sy.blockCache.Head().Number {
		sy.basevariable.SetMode(global.ModeNormal)
		sy.dc.Reset()
	} else {
		ilog.Infof("check sync process: now %v, end %v", sy.blockCache.Head().Number, sy.syncEnd)
	}
}

// OnBlockConfirmed confirms a block with block hash.
func (sy *SyncImpl) OnBlockConfirmed(hash string) {
	sy.dc.MissionComplete(hash)
}

// OnRecvBlock would free the peer.
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

func (sy *SyncImpl) getBlockHashes(start int64, end int64) *message.BlockHashResponse {
	resp := &message.BlockHashResponse{
		BlockHashes: make([]*message.BlockHash, 0, end-start+1),
	}
	node := sy.blockCache.Head()
	if node != nil && end > node.Number {
		end = node.Number
	}

	for i := end; i >= start; i-- {
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
	return resp
}

func (sy *SyncImpl) getBlockHashesByNums(nums []int64) *message.BlockHashResponse {
	resp := &message.BlockHashResponse{
		BlockHashes: make([]*message.BlockHash, 0, len(nums)),
	}
	var blk *block.Block
	var err error
	for _, num := range nums {
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
	return resp
}

func (sy *SyncImpl) handleHashQuery(rh *message.BlockHashQuery, peerID p2p.PeerID) {
	if rh.End < rh.Start || rh.Start < 0 {
		return
	}
	var resp *message.BlockHashResponse
	if rh.ReqType == 0 {
		resp = sy.getBlockHashes(rh.Start, rh.End)
	}
	if rh.ReqType == 1 {
		resp = sy.getBlockHashesByNums(rh.Nums)
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

// DownloadController defines the functions of download controller.
type DownloadController interface {
	OnRecvHash(hash string, peerID p2p.PeerID)
	OnTimeout(hash string, peerID p2p.PeerID)
	MissionComplete(hash string)
	FreePeer(hash string, peerID p2p.PeerID)
	Reset()
	Start()
	Stop()
}

const (
	// Done hash state type
	Done string = "Done"
	// Wait hash state type
	Wait string = "Wait"
)

const (
	peerConNum = 10
	// Free peer state type
	Free string = "Free"
)

const (
	// Head hashList node
	Head string = "Head"
	// Tail hashList node
	Tail string = "Tail"
)

type timerMap = map[string]*time.Timer

type hashListNode struct {
	val  string
	prev *hashListNode
	next *hashListNode
}

//DownloadControllerImpl is the implementation of DownloadController.
type DownloadControllerImpl struct {
	hashState      *sync.Map
	peerState      *sync.Map
	peerStateMutex *sync.Map
	peerMap        *sync.Map
	peerMapMutex   *sync.Map
	newPeerMutex   *sync.Mutex
	callback       callbackfunc
	chDownload     chan struct{}
	exitSignal     chan struct{}
}

// NewDownloadController returns a DownloadController instance.
func NewDownloadController(callback callbackfunc) (*DownloadControllerImpl, error) {
	dc := &DownloadControllerImpl{
		hashState:      new(sync.Map), // map[string]string
		peerState:      new(sync.Map), // map[PeerID](map[string]bool)
		peerStateMutex: new(sync.Map), // map[PeerID](metux)
		peerMap:        new(sync.Map), // map[PeerID](map[string]bool)
		peerMapMutex:   new(sync.Map), // map[PeerID](metux)
		newPeerMutex:   new(sync.Mutex),
		chDownload:     make(chan struct{}, 2),
		exitSignal:     make(chan struct{}),
		callback:       callback,
	}
	return dc, nil
}

// Reset resets data.
func (dc *DownloadControllerImpl) Reset() {
	dc.newPeerMutex.Lock()
	dc.hashState = new(sync.Map)
	dc.peerState = new(sync.Map)
	dc.peerStateMutex = new(sync.Map)
	dc.peerMap = new(sync.Map)
	dc.peerMapMutex = new(sync.Map)
	dc.newPeerMutex.Unlock()
}

// Start starts the DownloadController.
func (dc *DownloadControllerImpl) Start() {
	go dc.downloadLoop()
}

// Stop stops the DownloadController.
func (dc *DownloadControllerImpl) Stop() {
	close(dc.exitSignal)
}

func (dc *DownloadControllerImpl) getPeerMapMutex(peerID p2p.PeerID) (*sync.Mutex, bool) {
	pmMutexIF, ok := dc.peerMapMutex.Load(peerID)
	if !ok {
		ilog.Error("load peerMapMutex error")
		return nil, false
	}
	pmMutex, ok := pmMutexIF.(*sync.Mutex)
	if !ok {
		ilog.Error("change peerMapMutex error")
		return nil, false
	}
	return pmMutex, true
}

func (dc *DownloadControllerImpl) getStateMutex(peerID p2p.PeerID) (*sync.Mutex, bool) {
	psMutexIF, ok := dc.peerStateMutex.Load(peerID)
	if !ok {
		ilog.Error("load peerStateMutex error")
		return nil, false
	}
	psMutex, ok := psMutexIF.(*sync.Mutex)
	if !ok {
		ilog.Error("change peerStateMutex error")
		return nil, false
	}
	return psMutex, true
}

func (dc *DownloadControllerImpl) getHashMap(peerID p2p.PeerID) (*sync.Map, bool) {
	hmIF, ok := dc.peerMap.Load(peerID)
	if !ok {
		ilog.Error("load peerMap error")
		return nil, false
	}
	hashMap, ok := hmIF.(*sync.Map)
	if !ok {
		ilog.Error("change peerMap error")
		return nil, false
	}
	return hashMap, true
}

func (dc *DownloadControllerImpl) getHashListNode(hashMap *sync.Map, key string) (*hashListNode, bool) {
	nodeIF, ok := hashMap.Load(key)
	if !ok {
		ilog.Error("load tail node error")
		return nil, false
	}
	node, ok := nodeIF.(*hashListNode)
	if !ok {
		ilog.Error("change tail node error")
		return nil, false
	}
	return node, true
}

// OnRecvHash adds a mission.
func (dc *DownloadControllerImpl) OnRecvHash(hash string, peerID p2p.PeerID) {
	// ilog.Debugf("peer: %s, hash: %s", peerID, hash)
	hStateIF, _ := dc.hashState.LoadOrStore(hash, Wait)

	dc.newPeerMutex.Lock()
	_, ok := dc.peerState.Load(peerID)
	if !ok {
		pState := make(timerMap)
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
	dc.newPeerMutex.Unlock()
	if hashMap, ok := dc.getHashMap(peerID); ok {
		if _, ok = hashMap.Load(hash); !ok {
			pmMutex, ok := dc.getPeerMapMutex(peerID)
			if !ok {
				return
			}
			tail, ok := dc.getHashListNode(hashMap, Tail)
			if !ok {
				return
			}
			pmMutex.Lock()
			node := &hashListNode{val: hash, prev: tail.prev, next: tail}
			node.prev.next = node
			node.next.prev = node
			hashMap.Store(node.val, node)
			pmMutex.Unlock()
		}
	}
	if hState, ok := hStateIF.(string); ok && hState == Wait {
		select {
		case dc.chDownload <- struct{}{}:
		default:
		}
	}
}

// OnTimeout changes the hash state and frees the peer.
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
		psMutex, ok := dc.getStateMutex(peerID)
		if ok {
			psMutex.Lock()
			pState, ok := pStateIF.(timerMap)
			if !ok {
				ilog.Errorf("get peerstate error: %s", peerID.Pretty())
				// dc.peerState.Delete(peerID)
			} else {
				if _, ok = pState[hash]; ok {
					delete(pState, hash)
					select {
					case dc.chDownload <- struct{}{}:
					default:
					}
				}
			}
			psMutex.Unlock()
		}
	}
}

// MissionComplete changes the hash state.
func (dc *DownloadControllerImpl) MissionComplete(hash string) {
	dc.hashState.Store(hash, Done)
}

// FreePeer frees the peer.
func (dc *DownloadControllerImpl) FreePeer(hash string, peerID p2p.PeerID) {
	if pStateIF, ok := dc.peerState.Load(peerID); ok {
		psMutex, ok := dc.getStateMutex(peerID)
		if ok {
			psMutex.Lock()
			pState, ok := pStateIF.(timerMap)
			if !ok {
				ilog.Errorf("get peerstate error: %s", peerID.Pretty())
				// dc.peerState.Delete(peerID)
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
			psMutex.Unlock()
		}
	}
}

func (dc *DownloadControllerImpl) findWaitHashes(peerID p2p.PeerID, hashMap *sync.Map, ps timerMap, pmMutex *sync.Mutex, psMutex *sync.Mutex) {
	pmMutex.Lock()
	node, ok := dc.getHashListNode(hashMap, Head)
	if !ok {
		return
	}
	node = node.next
	pmMutex.Unlock()
	for {
		if node.val == Tail {
			return
		}
		hash := node.val
		var hState string
		hStateIF, ok := dc.hashState.Load(hash)
		if ok {
			hState, ok = hStateIF.(string)
		}
		if !ok || hState == Done {
			dc.hashState.Delete(hash)
			pmMutex.Lock()
			hashMap.Delete(hash)
			node.prev.next = node.next
			node.next.prev = node.prev
			pmMutex.Unlock()
		} else if hState == Wait {
			dc.hashState.Store(hash, peerID.Pretty())
			dc.callback(hash, peerID)
			psMutex.Lock()
			ps[hash] = time.AfterFunc(syncBlockTimeout, func() {
				dc.OnTimeout(hash, peerID)
			})
			psLen := len(ps)
			psMutex.Unlock()
			if psLen >= peerConNum {
				return
			}
		}
		pmMutex.Lock()
		node = node.next
		pmMutex.Unlock()
	}
}

func (dc *DownloadControllerImpl) downloadLoop() {
	for {
		select {
		case <-time.After(2 * syncBlockTimeout):
			select {
			case dc.chDownload <- struct{}{}:
			default:
			}
		case <-dc.chDownload:
			ilog.Debugf("Download Begin")
			dc.peerState.Range(func(k, v interface{}) bool {
				peerID := k.(p2p.PeerID)
				ilog.Debugf("peerID: %s", peerID.Pretty())
				ps, ok := v.(timerMap)
				if !ok {
					ilog.Errorf("get peerstate error: %s", peerID.Pretty())
					return true
				}
				pmMutex, pmmok := dc.getPeerMapMutex(peerID)
				psMutex, psmok := dc.getStateMutex(peerID)
				hashMap, hmok := dc.getHashMap(peerID)
				if !psmok || !pmmok || !hmok {
					return true
				}
				psMutex.Lock()
				ilog.Debugf("peerNum: %v", len(ps))
				psLen := len(ps)
				psMutex.Unlock()
				if psLen >= peerConNum {
					return true
				}
				dc.findWaitHashes(peerID, hashMap, ps, pmMutex, psMutex)
				return true
			})
			ilog.Debugf("Download End")
		case <-dc.exitSignal:
			return
		}
	}
}
