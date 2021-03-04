package synchro

import (
	"sync"
	"time"

	"github.com/iost-official/go-iost/v3/chainbase"
	"github.com/iost-official/go-iost/v3/ilog"
)

// rangeController will control the sync range.
type rangeController struct {
	start int64
	mutex *sync.RWMutex

	head  int64
	cBase *chainbase.ChainBase

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newRangeController(cBase *chainbase.ChainBase) *rangeController {
	r := &rangeController{
		start: 0,
		mutex: new(sync.RWMutex),

		head:  0,
		cBase: cBase,

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
	// ilog.Debug("sync range ", r.start, " to ", r.start+maxSyncRange-1)
	return r.start, r.start + maxSyncRange - 1
}

func (r *rangeController) setStart(start int64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.start = start
}

func (r *rangeController) updateStart() {
	head := r.cBase.HeadBlock().Head.Number
	lib := r.cBase.LIBlock().Head.Number
	if head > r.head {
		// Normal case
		r.head = head
		if lib+1 < r.head-maxSyncRange/2 {
			// TODO: This affects the maximum synchronization speed.
			// Sync End: head+500 is not enough. head+1000 is ok.
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
