package sync

import (
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"sync"
)

type requestHandler struct {
	p      p2p.Service
	bCache blockcache.BlockCache
	bChain *block.BlockChain

	requestCh chan p2p.IncomingMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newRequestHandler(p p2p.Service, bCache blockcache.BlockCache, bChain *block.BlockChain) *requestHandler {
	rHandler := &requestHandler{
		p:      p,
		bCache: bCache,
		bChain: bChain,

		requestCh: p.Register("sync request", p2p.SyncBlockHashRequest, p2p.SyncBlockRequest),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	rHandler.done.Add(1)
	go rHandler.controller()

	return rHandler
}

// Close will close the sync request handler.
func (r *requestHandler) Close() {
	close(r.quitCh)
	r.done.Wait()
	ilog.Infof("Stopped sync request handler.")
}

func (r *requestHandler) controller() {
	for {
		select {
		case request := <-r.requestCh:
			switch request.Type() {
			case p2p.SyncBlockHashRequest:
				r.handleBlockHashRequest(&request)
			case p2p.SyncBlockRequest:
				r.handleBlockRequest(&request)
			default:
				ilog.Warnf("Unexcept request type: %v", request.Type())
			}
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

func (r *requestHandler) handleBlockHashRequest(request *p2p.IncomingMessage) {
}

func (r *requestHandler) handleBlockRequest(request *p2p.IncomingMessage) {
}
