package sync

import (
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"sync"
)

const leastNeighborNumber = 5

type heightSync struct {
	p      p2p.Service
	bCache blockcache.BlockCache
	height int64

	msgCh chan p2p.IncomingMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newHeightSync(p p2p.Service, bCache blockcache.BlockCache) *heightSync {
	h := &heightSync{
		p:      p,
		bCache: bCache,
		height: 0,

		msgCh: p.Register("sync height response", p2p.SyncHeight),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	h.done.Add(1)
	go h.controller()

	return h
}

// Close will close the height synchronizer of blockchain.
func (h *heightSync) Close() {
	close(h.quitCh)
	h.done.Wait()
	ilog.Infof("Stopped height sync.")
}

// GetHeight will get the median of the head height of the neighbor nodes.
// If the number of neighbor nodes is less than leastNeighborNumber, return 0.
func (h *heightSync) GetHeight() {

}

func (h *heightSync) controller() {
	for {
		select {
		case msg := <-h.msgCh:
		default:
		}
		select {
		case <-r.quitCh:
			r.done.Done()
			return
		default:
		}
	}
}
