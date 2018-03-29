package protocol

import (
	"IOS/src/iosbase"
)

type ReqType int

const (
	ReqPrePrepare     ReqType = iota
	ReqPrepare
	ReqCommit
	ReqPublishTx
	ReqApplyBlockHash
	ReqApplyBlock
	ReqSendBlock
)

type ResState int

const (
	Accepted ResState = iota
	Reject
	Error
)

type NetworkFilter struct {
	base iosbase.Network
	*RuntimeData

	replicaChan  chan iosbase.Request
	recorderChan chan iosbase.Request
}

func (n *NetworkFilter) send(request iosbase.Request) chan iosbase.Response {
	return n.base.Send(request)
}

func (n *NetworkFilter) init(rd *RuntimeData, nw iosbase.Network) error {
	n.RuntimeData = rd
	n.base = nw

	n.replicaChan = make(chan iosbase.Request)
	n.recorderChan = make(chan iosbase.Request)

	return nil
}

func (n *NetworkFilter) router(reqChan chan iosbase.Request) {
	for n.isRunning {
		req := <-reqChan
		switch req.ReqType {
		case int(ReqPrePrepare):
			fallthrough
		case int(ReqPrepare):
			fallthrough
		case int(ReqCommit):
			n.replicaChan <- req
		case int(ReqPublishTx):
			fallthrough
		case int(ReqApplyBlockHash):
			fallthrough
		case int(ReqApplyBlock):
			fallthrough
		case int(ReqSendBlock):
			n.recorderChan <- req
		}
	}
}

func (n *NetworkFilter) replicaFilter(replica Replica, res chan iosbase.Response) {
	var req iosbase.Request
	for n.isRunning {
		req = <-n.replicaChan
		// 1. if req comes from right member
		if !n.view.isPrimary(req.From) && !n.view.isBackup(req.From) {
			res <- authorityError(req)
			continue
		}
		// 2. if req in right phase
		switch n.phase {
		case StartPhase:
			res <- invalidPhase(req)
		case PrePreparePhase:
			if req.ReqType == int(ReqPrePrepare) {
				res <- accept(req)
				replica.OnRequest(req)
			} else {
				res <- invalidPhase(req)
			}
		case PreparePhase:
			if req.ReqType == int(ReqPrepare) {
				res <- accept(req)
				replica.OnRequest(req)
			} else {
				res <- invalidPhase(req)
			}
		case CommitPhase:
			if req.ReqType == int(ReqCommit) {
				res <- accept(req)
				replica.OnRequest(req)
			} else {
				res <- invalidPhase(req)
			}
		case PanicPhase:
			res <- internalError(req)
		case EndPhase:
			res <- invalidPhase(req)
		default:
			res <- internalError(req)
		}

	}
}

func (n *NetworkFilter) recorderFilter(recorder Recorder, resChan chan iosbase.Response) {
	for n.isRunning {
		req := <-n.recorderChan
		switch req.ReqType {
		case int(ReqPublishTx):
			var tx iosbase.Tx
			err := tx.Decode(req.Body)
			if err != nil {
				resChan <- illegalTx(req)
			}
			recorder.PublishTx(tx)
		case int(ReqApplyBlockHash):
			recorder.OnAppliedBlockHash(req.From, resChan)
		case int(ReqApplyBlock):
			recorder.OnAppliedBlock(req.From)
		case int(ReqSendBlock):
			var blk iosbase.Block
			blk.Decode(req.Body)
			recorder.OnReceiveBlock(&blk)
		default:
			resChan <- internalError(req)
		}
	}
}

func invalidPhase(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Reject),
		Description: "Error: Invalid phase",
	}
}

func accept(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Accepted),
		Description: "",
	}
}

func internalError(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Reject),
		Description: "Error: Internal error",
	}
}

func authorityError(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Reject),
		Description: "Error: Authority error",
	}
}

func illegalTx(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Error),
		Description: "ERROR: Illegal Transaction",
	}
}
