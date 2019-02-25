package synchronizer

import (
	"sync"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	peer "github.com/libp2p/go-libp2p-peer"
)

// DownloadController defines the functions of download controller.
type DownloadController interface {
	CreateMission(hash string, p interface{}, peerID p2p.PeerID)
	Start()
	Stop()
	ReStart()
	DownloadLoop(mFunc MissionFunc)
	FreePeerLoop(fpFunc FreePeerFunc)
	FreePeer(hash string, peerID interface{})
	StopTimeout(hash string, peerID peer.ID)
}

const (
	// Done hash state type
	Done string = "Done"
	// Wait hash state type
	Wait string = "Wait"
)

const (
	syncBlockTimeout = 2 * time.Second
	peerConNum       = 15
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

// FreePeerFunc checks if the mission is completed.
type FreePeerFunc = func(hash string, p interface{}) (missionCompleted bool)

// MissionFunc checks if the mission is completed or tries to do the mission.
type MissionFunc = func(hash string, p interface{}, peerID interface{}) (missionAccept bool, missionCompleted bool)

type mapEntry struct {
	val  string
	p    interface{}
	prev *mapEntry
	next *mapEntry
}

// DownloadControllerImpl is the implementation of DownloadController.
type DownloadControllerImpl struct {
	hashState      *sync.Map
	peerState      *sync.Map
	peerStateMutex *sync.Map
	peerMap        *sync.Map
	peerMapMutex   *sync.Map
	newPeerMutex   *sync.Mutex
	fpFunc         FreePeerFunc
	mFunc          MissionFunc
	wg             *sync.WaitGroup
	chDownload     chan struct{}
	exitSignal     chan struct{}
}

// NewDownloadController returns a DownloadController instance.
func NewDownloadController(fpf FreePeerFunc, mf MissionFunc) (*DownloadControllerImpl, error) {
	dc := &DownloadControllerImpl{
		hashState:      new(sync.Map), // map[string]string
		peerState:      new(sync.Map), // map[PeerID](map[string]bool)
		peerStateMutex: new(sync.Map), // map[PeerID](metux)
		peerMap:        new(sync.Map), // map[PeerID](map[string]bool)
		peerMapMutex:   new(sync.Map), // map[PeerID](metux)
		newPeerMutex:   new(sync.Mutex),
		chDownload:     make(chan struct{}, 2),
		exitSignal:     make(chan struct{}),
		fpFunc:         fpf,
		mFunc:          mf,
		wg:             new(sync.WaitGroup),
	}
	return dc, nil
}

// ReStart restarts data.
func (dc *DownloadControllerImpl) ReStart() {
	dc.Stop()
	dc.newPeerMutex.Lock()
	dc.hashState = new(sync.Map)
	dc.peerState = new(sync.Map)
	dc.peerStateMutex = new(sync.Map)
	dc.peerMap = new(sync.Map)
	dc.peerMapMutex = new(sync.Map)
	dc.exitSignal = make(chan struct{})
	dc.wg = new(sync.WaitGroup)
	dc.newPeerMutex.Unlock()
	dc.Start()
}

// Stop stops the DownloadController.
func (dc *DownloadControllerImpl) Stop() {
	close(dc.exitSignal)
	dc.wg.Wait()
}

// Start starts the DownloadController.
func (dc *DownloadControllerImpl) Start() {
	dc.wg.Add(2)
	go dc.FreePeerLoop(dc.fpFunc)
	go dc.DownloadLoop(dc.mFunc)
}

func (dc *DownloadControllerImpl) getPeerMapMutex(key interface{}) (*sync.Mutex, bool) {
	pmMutexIF, ok := dc.peerMapMutex.Load(key)
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

func (dc *DownloadControllerImpl) getStateMutex(key interface{}) (*sync.Mutex, bool) {
	psMutexIF, ok := dc.peerStateMutex.Load(key)
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

func (dc *DownloadControllerImpl) getHashMap(key interface{}) (*sync.Map, bool) {
	hmIF, ok := dc.peerMap.Load(key)
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

func (dc *DownloadControllerImpl) getMapEntry(hashMap *sync.Map, key interface{}) (*mapEntry, bool) {
	nodeIF, ok := hashMap.Load(key)
	if !ok {
		ilog.Error("load tail node error")
		return nil, false
	}
	node, ok := nodeIF.(*mapEntry)
	if !ok {
		ilog.Error("change tail node error")
		return nil, false
	}
	return node, true
}

// CreateMission adds a mission.
func (dc *DownloadControllerImpl) CreateMission(hash string, p interface{}, peerID p2p.PeerID) {
	ilog.Debugf("create mission peer: %s, hash: %s, num: %v", peerID, common.Base58Encode([]byte(hash)), p.(int64))

	dc.newPeerMutex.Lock()
	if _, ok := dc.peerState.Load(peerID); !ok {
		pState := make(timerMap)
		pmMutex, _ := dc.peerMapMutex.LoadOrStore(peerID, new(sync.Mutex))
		hm, ok := dc.peerMap.LoadOrStore(peerID, new(sync.Map))
		if !ok {
			pmMutex.(*sync.Mutex).Lock()
			head := &mapEntry{val: Head, prev: nil, next: nil}
			tail := &mapEntry{val: Tail, prev: nil, next: nil}
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
			pmMutex.Lock()
			if _, ok = hashMap.Load(hash); !ok {
				tail, ok := dc.getMapEntry(hashMap, Tail)
				if ok {
					node := &mapEntry{val: hash, p: p, prev: tail.prev, next: tail}
					node.prev.next = node
					node.next.prev = node
					hashMap.Store(node.val, node)

					hState, _ := dc.hashState.LoadOrStore(hash, Wait)
					if hState == Wait {
						select {
						case dc.chDownload <- struct{}{}:
						default:
						}
					}
				}
			}
			pmMutex.Unlock()
		}
	}
}

func (dc *DownloadControllerImpl) missionComplete(hash string) {
	if _, ok := dc.hashState.Load(hash); ok {
		dc.hashState.Store(hash, Done)
	}
}

// FreePeer to free the peer and the mission
func (dc *DownloadControllerImpl) FreePeer(hash string, peerID interface{}) {
	if pStateIF, ok := dc.peerState.Load(peerID); ok {
		psMutex, ok := dc.getStateMutex(peerID)
		if ok {
			psMutex.Lock()
			pState, ok := pStateIF.(timerMap)
			if !ok {
				ilog.Errorf("get peerstate error: %s", peerID.(p2p.PeerID).Pretty())
				//dc.peerState.Delete(peerID)
			} else {
				if timer, ok := pState[hash]; ok {
					timer.Stop()
					delete(pState, hash)
					ilog.Debugf("free peer, peerID:%s", peerID.(p2p.PeerID).Pretty())
					select {
					case dc.chDownload <- struct{}{}:
					default:
					}
				}
			}
			psMutex.Unlock()
		}
	}
	if hState, ok := dc.hashState.Load(hash); ok {
		if hState == peerID {
			dc.hashState.Store(hash, Wait)
		}
	}
}

func (dc *DownloadControllerImpl) handleFreePeer(fpFunc FreePeerFunc) {
	ilog.Debugf("free peer begin")
	dc.peerState.Range(func(peerID, v interface{}) bool {
		ps, ok := v.(timerMap)
		if !ok {
			ilog.Errorf("get peerstate error: %s", peerID.(p2p.PeerID).Pretty())
		}
		psMutex, psmok := dc.getStateMutex(peerID)
		hashMap, hmok := dc.getHashMap(peerID)
		if !psmok || !hmok {
			return true
		}
		psMutex.Lock()
		hashlist := make([]string, 0, len(ps))
		for hash := range ps {
			hashlist = append(hashlist, hash)
		}
		psMutex.Unlock()
		for _, hash := range hashlist {
			hState, ok := dc.hashState.Load(hash)
			if ok {
				if hState == Done || hState == Wait {
					psMutex.Lock()
					delete(ps, hash)
					psMutex.Unlock()
					select {
					case dc.chDownload <- struct{}{}:
					default:
					}
				} else {
					var node *mapEntry
					nodeIF, ok := hashMap.Load(hash)
					if ok {
						node, ok = nodeIF.(*mapEntry)
					}
					if ok && fpFunc(hash, node.p) {
						dc.FreePeer(hash, peerID)
					}
				}
			}
		}
		return true
	})
	ilog.Debugf("free peer end")
}

// FreePeerLoop is the Loop to free the peer.
func (dc *DownloadControllerImpl) FreePeerLoop(fpFunc FreePeerFunc) {
	defer dc.wg.Done()
	checkPeerTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-checkPeerTicker.C:
			dc.handleFreePeer(fpFunc)
		case <-dc.exitSignal:
			return
		}
	}
}

func (dc *DownloadControllerImpl) handleDownload(peerID interface{}, hashMap *sync.Map, ps timerMap, pmMutex *sync.Mutex, psMutex *sync.Mutex, mFunc MissionFunc) bool {
	psMutex.Lock()
	ilog.Debugf("peerNum: %v", len(ps))
	psLen := len(ps)
	psMutex.Unlock()
	if psLen >= peerConNum {
		return true
	}
	pmMutex.Lock()
	node, ok := dc.getMapEntry(hashMap, Head)
	if !ok {
		return true
	}
	node = node.next
	pmMutex.Unlock()
	for {
		if node.val == Tail {
			return true
		}
		hash := node.val
		hState, ok := dc.hashState.Load(hash)
		if !ok || hState == Done {
			dc.hashState.Delete(hash)
			pmMutex.Lock()
			hashMap.Delete(hash)
			node.prev.next = node.next
			node.next.prev = node.prev
			pmMutex.Unlock()
		} else if hState == Wait {
			mok, mdone := mFunc(hash, node.p, peerID)
			if mok {
				dc.hashState.Store(hash, peerID)
				psMutex.Lock()
				ps[hash] = time.AfterFunc(syncBlockTimeout, func() {
					ilog.Debugf("sync timout, hash=%v, peerID=%s", common.Base58Encode([]byte(hash)), peerID.(p2p.PeerID).Pretty())
					dc.FreePeer(hash, peerID)
				})
				psLen := len(ps)
				psMutex.Unlock()
				if psLen >= peerConNum {
					return true
				}
			}
			if mdone {
				dc.missionComplete(hash)
			}
		}
		pmMutex.Lock()
		node = node.next
		pmMutex.Unlock()
	}
}

// DownloadLoop is the Loop to download the mission.
func (dc *DownloadControllerImpl) DownloadLoop(mFunc MissionFunc) {
	defer dc.wg.Done()
	for {
		select {
		case <-time.After(2 * syncBlockTimeout):
			select {
			case dc.chDownload <- struct{}{}:
			default:
			}
		case <-dc.chDownload:
			ilog.Debugf("Download Begin")
			dc.peerState.Range(func(peerID, v interface{}) bool {
				select {
				case <-dc.exitSignal:
					return false
				default:
					ilog.Debugf("peerID: %s", peerID.(p2p.PeerID).Pretty())
					ps, ok := v.(timerMap)
					if !ok {
						ilog.Errorf("get peerstate error: %s", peerID.(p2p.PeerID).Pretty())
						return true
					}
					pmMutex, pmmok := dc.getPeerMapMutex(peerID)
					psMutex, psmok := dc.getStateMutex(peerID)
					hashMap, hmok := dc.getHashMap(peerID)
					if !psmok || !pmmok || !hmok {
						return true
					}

					dc.handleDownload(peerID, hashMap, ps, pmMutex, psMutex, mFunc)
					return true
				}
			})
			ilog.Debugf("Download End")
		case <-dc.exitSignal:
			return
		}
	}
}

// StopTimeout make the timer stop.
func (dc *DownloadControllerImpl) StopTimeout(hash string, peerID peer.ID) {
	psIF, ok := dc.peerState.Load(peerID)
	if !ok {
		return
	}
	ps, ok := psIF.(timerMap)
	if !ok {
		return
	}
	psMutex, psmok := dc.getStateMutex(peerID)
	if !psmok {
		return
	}
	psMutex.Lock()
	defer psMutex.Unlock()
	if timer, ok := ps[hash]; ok {
		timer.Stop()
	}
}
