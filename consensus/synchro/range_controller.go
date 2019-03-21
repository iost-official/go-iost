package synchro

import (
	"sync"
	"time"

	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/ilog"
)

// rangeController will control the sync range.
type rangeController struct {
	start int64
	mutex *sync.RWMutex

	head   int64
	bCache blockcache.BlockCache

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newRangeController(bCache blockcache.BlockCache) *rangeController {
	r := &rangeController{
		start: 0,
		mutex: new(sync.RWMutex),

		head:   0,
		bCache: bCache,

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	r.done.Add(1)
	go r.controller()

	return r
}

// Close will close the rangeController.
func (r *rangeController) Close() {
	close(r.quitCh)
	r.done.Wait()
	ilog.Infof("Stopped range controller.")
}

// SyncRange return the range of synchronization required.
func (r *rangeController) SyncRange() (start int64, end int64) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if r.start < 0 {
		return 0, maxSyncRange - 1
	}
	return r.start, r.start + maxSyncRange - 1
}

func (r *rangeController) setStart(start int64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.start = start
}

func (r *rangeController) updateStart() {
	head := r.bCache.Head().Head.Number
	lib := r.bCache.LinkedRoot().Head.Number
	if head > r.head {
		// Normal case
		r.head = head
		if lib+1 < r.head-maxSyncRange/2 {
			r.setStart(r.head - maxSyncRange/2)
		} else {
			r.setStart(lib + 1)
		}
	} else {
		// When the network does not reach a consensus for a long time.
		r.setStart(lib + 1)
		for r.start < r.head-maxSyncRange/2 {
			time.Sleep(2 * time.Second)
			r.setStart(r.start + maxSyncRange/10)
		}
	}
}

func (r *rangeController) controller() {
	for {
		select {
		case <-time.After(2 * time.Second):
			r.updateStart()
		case <-r.quitCh:
			r.done.Done()
			return
		}
	}
}
