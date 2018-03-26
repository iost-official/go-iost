package protocol

import "IOS/src/iosbase"

type NetworkFilter struct {
	base iosbase.Network

	receiveChan chan iosbase.Request
}

func (c *Consensus) send (request iosbase.Request) {
	c.base.Send(request)
}

func (c *Consensus) networkFilterInit(nw iosbase.Network) error {
	c.base = nw
	return nil
}


func (c *Consensus) listenAndFilter() {
	var request iosbase.Request
	rawReq, rawRes, err := c.base.Listen(Port)
	if err != nil {
		panic(err)
	}
	for true {
		request = <-rawReq
		// 1. if request comes from right member
		if c.view.isPrimary(request.From) || c.view.isBackup(request.From) {
			rawRes <-iosbase.Response{request.To, request.From, int(Reject), "YOU ARE NOT MEMBER"}
		}
		// 2. if request in right phase
		switch c.phase {
		case StartPhase:
			rawRes <-iosbase.Response{request.To, request.From, int(Reject), "ON START PHASE"}
		case PrePreparePhase:
			if request.ReqType == int(PrePreparePhase) {
				rawRes <-iosbase.Response{request.To, request.From, int(Accepted), ""}
				c.receiveChan <- request
			} else {
				rawRes <-iosbase.Response{request.To, request.From, int(Reject), "ON START PHASE"}
			}
		case PreparePhase:
			if request.ReqType == int(PreparePhase) {
				rawRes <-iosbase.Response{request.To, request.From, int(Accepted), ""}
				c.receiveChan <- request
			} else {
				rawRes <-iosbase.Response{request.To, request.From, int(Reject), "ON START PHASE"}
			}
		case CommitPhase:
			if request.ReqType == int(CommitPhase) {
				rawRes <-iosbase.Response{request.To, request.From, int(Accepted), ""}
				c.receiveChan <- request
			} else {
				rawRes <-iosbase.Response{request.To, request.From, int(Reject), "ON START PHASE"}
			}
		case PanicPhase:
			rawRes <-iosbase.Response{request.To, request.From, int(Error), "INTERNAL ERROR"}
		case EndPhase:
			rawRes <-iosbase.Response{request.To, request.From, int(Reject), "ON START PHASE"}
		}
		rawRes <-iosbase.Response{request.To, request.From, int(Reject), "ON START PHASE"}
	}
}

type ResponseState int

const (
	Accepted ResponseState = iota
	Reject
	Error
)

