package synchro

import (
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/consensus/synchronizer/pb"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

const (
	leastNeighborNumber  = 5
	heightExpiredSeconds = 60
)

type heightSync struct {
	neighborHeight map[p2p.PeerID]*msgpb.SyncHeight
	mutex          *sync.RWMutex

	msgCh chan p2p.IncomingMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newHeightSync(p p2p.Service) *heightSync {
	h := &heightSync{
		neighborHeight: make(map[p2p.PeerID]*msgpb.SyncHeight),
		mutex:          new(sync.RWMutex),

		msgCh: p.Register("sync height response", p2p.SyncHeight),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	h.done.Add(2)
	go h.syncHeightController()
	go h.expirationController()

	return h
}

// Close will close the height synchronizer of blockchain.
func (h *heightSync) Close() {
	close(h.quitCh)
	h.done.Wait()
	ilog.Infof("Stopped height sync.")
}

// NeighborHeight will return the median of the head height of the neighbor nodes.
// If the number of neighbor nodes is less than leastNeighborNumber, return -1.
func (h *heightSync) NeighborHeight() int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if len(h.neighborHeight) < leastNeighborNumber {
		return -1
	}

	t := make([]int64, 0)
	for _, v := range h.neighborHeight {
		t = append(t, v.Height)
	}
	sort.Slice(t, func(i, j int) bool { return t[i] < t[j] })

	return t[len(t)/2]
}

func (h *heightSync) handleHeightSync(msg *p2p.IncomingMessage) {
	if msg.Type() != p2p.SyncHeight {
		ilog.Warnf("Expect the type %v, but get a unexpected type %v", p2p.SyncHeight, msg.Type())
		return
	}

	syncHeight := &msgpb.SyncHeight{}
	err := proto.Unmarshal(msg.Data(), syncHeight)
	if err != nil {
		ilog.Warnf("Unmarshal sync height failed: %v", err)
		return
	}

	ilog.Debugf("Received height %v from peer %v.", syncHeight.Height, msg.From().Pretty())

	h.mutex.Lock()
	defer h.mutex.Unlock()

	if old, ok := h.neighborHeight[msg.From()]; ok {
		if old.Time < syncHeight.Time {
			h.neighborHeight[msg.From()] = syncHeight
		}
	} else {
		h.neighborHeight[msg.From()] = syncHeight
	}
}

func (h *heightSync) syncHeightController() {
	for {
		select {
		case msg := <-h.msgCh:
			h.handleHeightSync(&msg)
		case <-h.quitCh:
			h.done.Done()
			return
		}
	}
}

func (h *heightSync) doExpiration() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	now := time.Now().Unix()
	for k, v := range h.neighborHeight {
		if v.Time+heightExpiredSeconds < now {
			delete(h.neighborHeight, k)
		}
	}
}

func (h *heightSync) expirationController() {
	for {
		select {
		case <-time.After(2 * time.Second):
			h.doExpiration()
		case <-h.quitCh:
			h.done.Done()
			return
		}
	}
}
