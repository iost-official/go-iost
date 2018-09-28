package synchronizer

import (
	"sync"
	"time"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

// DownloadController defines the functions of download controller.
type DownloadController interface {
	CreateMission(hash string, p interface{}, peerID p2p.PeerID)
	MissionTimeout(hash string, peerID p2p.PeerID)
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
	peerConNum = 5
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

type callbackfunc = func(hash string, p interface{}, peerID p2p.PeerID) bool

type hashListNode struct {
	val  string
	p    interface{}
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
func (dc *DownloadControllerImpl) CreateMission(hash string, p interface{}, peerID p2p.PeerID) {
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
			if _, ok = hashMap.Load(hash); !ok {
				tail, ok := dc.getHashListNode(hashMap, Tail)
				if !ok {
					return
				}
				pmMutex.Lock()
				node := &hashListNode{val: hash, p: p, prev: tail.prev, next: tail}
				node.prev.next = node
				node.next.prev = node
				hashMap.Store(node.val, node)
			}
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
func (dc *DownloadControllerImpl) MissionTimeout(hash string, peerID p2p.PeerID) {
	ilog.Debugf("sync timout, hash=%v, peerID=%s", []byte(hash), peerID.Pretty())
	if hStateIF, ok := dc.hashState.Load(hash); ok {
		hState, ok := hStateIF.(string)
		if !ok {
			ilog.Errorf("get hash state error: %s", hash)
			// dc.hashState.Delete(hash)
		} else if hState == peerID.Pretty() {
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
	if _, ok := dc.hashState.Load(hash); ok {
		dc.hashState.Store(hash, Done)
	}
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
	if hStateIF, ok := dc.hashState.Load(hash); ok {
		hState, ok := hStateIF.(string)
		if !ok {
			ilog.Errorf("get hash state error: %s", hash)
			// dc.hashState.Delete(hash)
		} else if hState == peerID.Pretty() {
			dc.hashState.Store(hash, Wait)
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
			if dc.callback(hash, node.p, peerID) {
				dc.hashState.Store(hash, peerID.Pretty())
				psMutex.Lock()
				ps[hash] = time.AfterFunc(syncBlockTimeout, func() {
					dc.MissionTimeout(hash, peerID)
				})
				psLen := len(ps)
				psMutex.Unlock()
				if psLen >= peerConNum {
					return
				}
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
