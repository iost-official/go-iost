package synchro

import (
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/iost-official/go-iost/v3/chainbase"
	"github.com/iost-official/go-iost/v3/common"
	msgpb "github.com/iost-official/go-iost/v3/consensus/synchro/pb"
	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/p2p"
	"google.golang.org/protobuf/proto"
)

var (
	workerPoolSize = 2
	timeout        = 2 * time.Second
)

// requestHandler is responsible for processing synchronization requests from other nodes.
type requestHandler struct {
	pool *tunny.Pool

	requestCh chan p2p.IncomingMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newRequestHandler(cBase *chainbase.ChainBase, p p2p.Service) *requestHandler {
	worker := newRequestHandlerWorker(cBase, p)
	rHandler := &requestHandler{
		pool: tunny.NewFunc(workerPoolSize, worker.process),

		requestCh: p.Register("sync request", p2p.SyncBlockHashRequest, p2p.SyncBlockRequest, p2p.NewBlockRequest),

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

func (r *requestHandler) handle(request *p2p.IncomingMessage) {
	_, err := r.pool.ProcessTimed(request, timeout)
	if err == tunny.ErrJobTimedOut {
		ilog.Warnf("The request %v from %v timed out", request.Type(), request.From().String())
	}
}

func (r *requestHandler) controller() {
	for {
		select {
		case request := <-r.requestCh:
			go r.handle(&request)
		case <-r.quitCh:
			r.done.Done()
			return
		}
	}
}

type requestHandlerWorker struct {
	cBase *chainbase.ChainBase
	p     p2p.Service
}

func newRequestHandlerWorker(cBase *chainbase.ChainBase, p p2p.Service) *requestHandlerWorker {
	r := &requestHandlerWorker{
		cBase: cBase,
		p:     p,
	}

	return r
}

func (r *requestHandlerWorker) getBlockHashResponse(start int64, end int64) (*msgpb.BlockHashResponse, int64) {
	blockInfos := make([]*msgpb.BlockInfo, 0)
	var waitNum int64 = 0
	for num := start; num <= end; num++ {
		hash, ok := r.cBase.GetBlockHashByNum(num)
		if !ok {
			if waitNum == 0 {
				waitNum = num
			}
			ilog.Debugf("Get block by num %v failed.", num)
			break
		}
		blockInfo := &msgpb.BlockInfo{
			Number: num,
			Hash:   hash,
		}
		blockInfos = append(blockInfos, blockInfo)
	}

	return &msgpb.BlockHashResponse{
		BlockInfos: blockInfos,
	}, waitNum
}

func (r *requestHandlerWorker) handleBlockHashRequest(request *p2p.IncomingMessage) {
	t1 := time.Now()
	blockHashQuery := &msgpb.BlockHashQuery{}
	if err := proto.Unmarshal(request.Data(), blockHashQuery); err != nil {
		ilog.Warnf("Unmarshal BlockHashQuery failed: %v", err)
		return
	}

	// RequireType_GETBLOCKHASHESBYNUMBER is deprecated.
	if blockHashQuery.ReqType == msgpb.RequireType_GETBLOCKHASHESBYNUMBER {
		return
	}

	start := blockHashQuery.Start
	end := blockHashQuery.End

	if (start < 0) || (start > end) || (end-start+1 > maxSyncRange) {
		ilog.Warnf("Receive attack request from peer %v, start: %v, end: %v.", request.From().String(), blockHashQuery.Start, blockHashQuery.End)
		return
	}

	head := r.cBase.HeadBlock().Head.Number
	// Because this request is broadcast, so there is this situation.
	// It will be changed later.
	if start > head {
		return
	}
	if end > head {
		end = head
	}
	blockHashResponse, waitNum := r.getBlockHashResponse(start, end)

	msg, err := proto.Marshal(blockHashResponse)
	if err != nil {
		ilog.Warnf("Marshal BlockHashResponse failed: struct=%+v, err=%v", blockHashResponse, err)
		return
	}
	t2 := time.Now()
	r.p.SendToPeer(request.From(), msg, p2p.SyncBlockHashResponse, p2p.NormalMessage)
	t3 := time.Now()
	if t3.Sub(t1) > 2*time.Second {
		ilog.Errorf("handleBlockHashRequest timeout. fetching time: %v, network time: %v, sync range: %v %v %v", t2.Sub(t1), t3.Sub(t2), start, waitNum, end)
	}
}

func (r *requestHandlerWorker) handleBlockRequest(request *p2p.IncomingMessage, mtype p2p.MessageType, priority p2p.MessagePriority) {
	blockInfo := &msgpb.BlockInfo{}
	if err := proto.Unmarshal(request.Data(), blockInfo); err != nil {
		ilog.Warnf("Unmarshal BlockInfo failed: %v", err)
		return
	}

	block, ok := r.cBase.GetBlockByHash(blockInfo.Hash)
	if !ok {
		ilog.Warnf("Handle block request failed, from=%v, hash=%v.", request.From().String(), common.Base58Encode(blockInfo.Hash))
		return
	}

	msg, err := block.Encode()
	if err != nil {
		ilog.Errorf("Encode block failed: %v\nblock: %+v", err, block)
		return
	}
	r.p.SendToPeer(request.From(), msg, mtype, priority)
}

func (r *requestHandlerWorker) process(payload any) any {
	request, ok := payload.(*p2p.IncomingMessage)
	if !ok {
		ilog.Warnf("Assert payload %+v to IncomingMessage failed", payload)
		return nil
	}

	switch request.Type() {
	case p2p.SyncBlockHashRequest:
		r.handleBlockHashRequest(request)
	case p2p.SyncBlockRequest:
		r.handleBlockRequest(request, p2p.SyncBlockResponse, p2p.NormalMessage)
	case p2p.NewBlockRequest:
		r.handleBlockRequest(request, p2p.NewBlock, p2p.UrgentMessage)
	default:
		ilog.Warnf("Unexcept request type: %v", request.Type())
	}

	return nil
}
