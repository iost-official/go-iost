package protocol

import "IOS/src/iosbase"

type NetworkFilter struct {
	base iosbase.Network
	*RuntimeData

	replicaChan  chan iosbase.Request
	recorderChan chan iosbase.Request
}

func (n *NetworkFilter) send(request iosbase.Request) {
	n.base.Send(request)
}

func (n *NetworkFilter) init(rd *RuntimeData, nw iosbase.Network) error {
	n.RuntimeData = rd
	n.base = nw
	return nil
}

type ResponseState int

const (
	Accepted ResponseState = iota
	Reject
	Error
)

func (n *NetworkFilter) replicaFilter(request chan iosbase.Request, res chan iosbase.Response) {
	var req iosbase.Request
	for n.isRunning {
		req = <-request
		// 1. if req comes from right member
		if n.view.isPrimary(req.From) || n.view.isBackup(req.From) {
			res <- authorityError(req)
		}
		// 2. if req in right phase
		switch n.phase {
		case StartPhase:
			res <- invalidPhase(req)
		case PrePreparePhase:
			if req.ReqType == int(PrePreparePhase) {
				res <- accept(req)
				n.replicaChan <- req
			} else {
				res <- invalidPhase(req)
			}
		case PreparePhase:
			if req.ReqType == int(PreparePhase) {
				res <- accept(req)
				n.replicaChan <- req
			} else {
				res <- invalidPhase(req)
			}
		case CommitPhase:
			if req.ReqType == int(CommitPhase) {
				res <- accept(req)
				n.replicaChan <- req
			} else {
				res <- invalidPhase(req)
			}
		case PanicPhase:
			res <- internalError(req)
		case EndPhase:
			res <- invalidPhase(req)
		}
		res <- internalError(req)
	}
}

func (n *NetworkFilter) recorderFilter(recorder Recorder, reqChan chan iosbase.Request, resChan chan iosbase.Response) {
	for n.isRunning {
		req := <-reqChan
		var tx iosbase.Tx
		err := tx.Decode(req.Body)
		if err != nil {
			resChan <- illegalTx(req)
		}
		recorder.publishTx(tx)
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
