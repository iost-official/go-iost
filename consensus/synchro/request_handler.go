package synchro

import (
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/synchro/pb"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

// requestHandler is responsible for processing synchronization requests from other nodes.
type requestHandler struct {
	p      p2p.Service
	bCache blockcache.BlockCache
	bChain block.Chain

	requestCh chan p2p.IncomingMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newRequestHandler(p p2p.Service, bCache blockcache.BlockCache, bChain block.Chain) *requestHandler {
	rHandler := &requestHandler{
		p:      p,
		bCache: bCache,
		bChain: bChain,

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

func (r *requestHandler) getBlockByHash(hash []byte) *block.Block {
	block, err := r.bCache.GetBlockByHash(hash)
	if err != nil {
		block, err := r.bChain.GetBlockByHash(hash)
		if err != nil {
			ilog.Warnf("Get block by hash %v failed: %v", common.Base58Encode(hash), err)
			return nil
		}
		return block
	}
	return block
}

func (r *requestHandler) getBlockHashResponse(start int64, end int64) *msgpb.BlockHashResponse {
	blockInfos := make([]*msgpb.BlockInfo, 0)
	for num := start; num <= end; num++ {
		// This code is ugly and then optimize the bCache and bChain.
		var hash []byte
		if blk, err := r.bCache.GetBlockByNumber(num); err != nil {
			hash, err = r.bChain.GetHashByNumber(num)
			if err != nil {
				ilog.Warnf("Get hash by num %v failed: %v", num, err)
				continue
			}
		} else {
			hash = blk.HeadHash()
		}
		blockInfo := &msgpb.BlockInfo{
			Number: num,
			Hash:   hash,
		}
		blockInfos = append(blockInfos, blockInfo)
	}

	return &msgpb.BlockHashResponse{
		BlockInfos: blockInfos,
	}
}

func (r *requestHandler) handleBlockHashRequest(request *p2p.IncomingMessage) {
	blockHashQuery := &msgpb.BlockHashQuery{}
	if err := proto.Unmarshal(request.Data(), blockHashQuery); err != nil {
		ilog.Warnf("Unmarshal BlockHashQuery failed: %v", err)
		return
	}

	// RequireType_GETBLOCKHASHESBYNUMBER is deprecated.
	if blockHashQuery.ReqType == msgpb.RequireType_GETBLOCKHASHESBYNUMBER {
		return
	}

	if (blockHashQuery.Start < 0) ||
		(blockHashQuery.Start > blockHashQuery.End) ||
		(blockHashQuery.End-blockHashQuery.Start+1 > maxSyncRange) {
		ilog.Warnf("Receive attack request from peer %v, start: %v, end: %v.", request.From().Pretty(), blockHashQuery.Start, blockHashQuery.End)
		return
	}

	// Because this request is broadcast, so there is this situation.
	// It will be changed later.
	if blockHashQuery.Start > r.bCache.Head().Head.Number {
		return
	}

	blockHashResponse := r.getBlockHashResponse(blockHashQuery.Start, blockHashQuery.End)

	msg, err := proto.Marshal(blockHashResponse)
	if err != nil {
		ilog.Warnf("Marshal BlockHashResponse failed: struct=%+v, err=%v", blockHashResponse, err)
		return
	}
	r.p.SendToPeer(request.From(), msg, p2p.SyncBlockHashResponse, p2p.NormalMessage)
}

func (r *requestHandler) handleBlockRequest(request *p2p.IncomingMessage, priority p2p.MessagePriority) {
	blockInfo := &msgpb.BlockInfo{}
	if err := proto.Unmarshal(request.Data(), blockInfo); err != nil {
		ilog.Warnf("Unmarshal BlockInfo failed: %v", err)
		return
	}

	block := r.getBlockByHash(blockInfo.Hash)
	if block == nil {
		ilog.Warnf("Handle block request failed, from=%v, hash=%v.", request.From().Pretty(), common.Base58Encode(blockInfo.Hash))
		return
	}

	msg, err := block.Encode()
	if err != nil {
		ilog.Errorf("Encode block failed: %v\nblock: %+v", err, block)
		return
	}
	r.p.SendToPeer(request.From(), msg, p2p.SyncBlockResponse, priority)
}

func (r *requestHandler) controller() {
	for {
		select {
		case request := <-r.requestCh:
			// TODO: Need a thread pool here.
			switch request.Type() {
			case p2p.SyncBlockHashRequest:
				go r.handleBlockHashRequest(&request)
			case p2p.SyncBlockRequest:
				go r.handleBlockRequest(&request, p2p.NormalMessage)
			case p2p.NewBlockRequest:
				go r.handleBlockRequest(&request, p2p.UrgentMessage)
			default:
				ilog.Warnf("Unexcept request type: %v", request.Type())
			}
		case <-r.quitCh:
			r.done.Done()
			return
		}
	}
}
