package protocol

import "IOS/src/iosbase"

type NetworkFilter struct {
	base iosbase.Network

	valiChan  chan iosbase.Request
	repliChan chan iosbase.Request
}

func (c *Consensus) send(request iosbase.Request) {
	c.base.Send(request)
}

func (c *Consensus) networkFilterInit(nw iosbase.Network) error {
	c.base = nw
	return nil
}


type ResponseState int

const (
	Accepted ResponseState = iota
	Reject
	Error
)

func (c *Consensus) replicaFilter(request chan iosbase.Request, res chan iosbase.Response) {
	var req iosbase.Request
	for c.isRunning {
		req = <-request
		// 1. if req comes from right member
		if c.view.isPrimary(req.From) || c.view.isBackup(req.From) {
			res <- iosbase.Response{req.To, req.From, int(Reject), "YOU ARE NOT MEMBER"}
		}
		// 2. if req in right phase
		switch c.phase {
		case StartPhase:
			res <- iosbase.Response{req.To, req.From, int(Reject), "ON START PHASE"}
		case PrePreparePhase:
			if req.ReqType == int(PrePreparePhase) {
				res <- iosbase.Response{req.To, req.From, int(Accepted), ""}
				c.valiChan <- req
			} else {
				res <- iosbase.Response{req.To, req.From, int(Reject), "ON START PHASE"}
			}
		case PreparePhase:
			if req.ReqType == int(PreparePhase) {
				res <- iosbase.Response{req.To, req.From, int(Accepted), ""}
				c.valiChan <- req
			} else {
				res <- iosbase.Response{req.To, req.From, int(Reject), "ON START PHASE"}
			}
		case CommitPhase:
			if req.ReqType == int(CommitPhase) {
				res <- iosbase.Response{req.To, req.From, int(Accepted), ""}
				c.valiChan <- req
			} else {
				res <- iosbase.Response{req.To, req.From, int(Reject), "ON START PHASE"}
			}
		case PanicPhase:
			res <- iosbase.Response{req.To, req.From, int(Error), "INTERNAL ERROR"}
		case EndPhase:
			res <- iosbase.Response{req.To, req.From, int(Reject), "ON START PHASE"}
		}
		res <- iosbase.Response{req.To, req.From, int(Reject), "ON START PHASE"}
	}
}

func (c *Consensus) recorderFilter(req chan iosbase.Request, res chan iosbase.Response) {
	for c.isRunning {
		request := <-req
		var tx iosbase.Tx
		tx.Decode(request.Body)
		c.PublishTx(tx)
	}
}
