package protocol

import (
	"fmt"
	"github.com/iost-official/PrototypeWorks/iosbase"
)

/*
Marked request types using by protocol
*/
type ReqType int

const (
	ReqPrePrepare ReqType = iota
	ReqPrepare
	ReqCommit
	ReqSubmitTxPack
	ReqPublishTx
	ReqNewBlock
)

type ResState int

const (
	Accepted ResState = iota
	Reject
	Error
)

//go:generate mockgen -destination mocks/mock_router.go -package protocol_mock github.com/iost-official/PrototypeWorks/protocol Router

/*
Forwarding specific request to other components and sending messages for them
*/
type Router interface {
	Init(base iosbase.Network, port uint16) error
	FilteredChan(filter Filter) (chan iosbase.Request, chan iosbase.Response, error)
	Run()
	Stop()
	Send(req iosbase.Request) chan iosbase.Response
	Broadcast(req iosbase.Request)
}

func RouterFactory(target string) (Router, error) {
	switch target {
	case "base":
		return &RouterImpl{}, nil
	}
	return nil, fmt.Errorf("target Router not found")
}

type RouterImpl struct {
	base iosbase.Network

	chIn, chOut chan iosbase.Request
	chReply     chan iosbase.Response

	filterList  []Filter
	filterMap   map[int]chan iosbase.Request
	knownMember []string
	ExitSignal  chan bool
}

func (r *RouterImpl) Init(base iosbase.Network, port uint16) error {
	var err error

	r.base = base
	r.filterList = make([]Filter, 0)
	r.filterMap = make(map[int]chan iosbase.Request)
	r.knownMember = make([]string, 0)
	r.ExitSignal = make(chan bool)

	r.chIn, r.chReply, err = r.base.Listen(port)
	if err != nil {
		return err
	}

	return nil
}

/*
Get filtered request channel
*/
func (r *RouterImpl) FilteredChan(filter Filter) (chan iosbase.Request, chan iosbase.Response, error) {
	chReq := make(chan iosbase.Request)

	r.filterList = append(r.filterList, filter)
	r.filterMap[len(r.filterList)-1] = chReq

	return chReq, r.chReply, nil
}

func (r *RouterImpl) receiveLoop() {
	for true {
		select {
		case <-r.ExitSignal:
			return
		case req := <-r.chIn:
			for i, f := range r.filterList {
				if f.check(req) {
					r.filterMap[i] <- req
				}
			}
		}
	}
}

func (r *RouterImpl) Run() {
	go r.receiveLoop()

}
func (r *RouterImpl) Stop() {
	r.ExitSignal <- true
}
func (r *RouterImpl) Send(req iosbase.Request) chan iosbase.Response {
	return r.base.Send(req)
}

func (r *RouterImpl) Broadcast(req iosbase.Request) {
	for _, to := range r.knownMember {
		req.To = to

		go func() {
			r.Send(req)
		}()
	}
}

func invalidPhase(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		Time:        req.Time,
		From:        req.To,
		To:          req.From,
		Code:        int(Reject),
		Description: "Error: Invalid phase",
	}
}

func accept(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		Time:        req.Time,
		From:        req.To,
		To:          req.From,
		Code:        int(Accepted),
		Description: "Accepted",
	}
}

func internalError(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		Time:        req.Time,
		From:        req.To,
		To:          req.From,
		Code:        int(Reject),
		Description: "Error: Internal error",
	}
}

func authorityError(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		Time:        req.Time,
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

func syntaxError(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		Time:        req.Time,
		From:        req.To,
		To:          req.From,
		Code:        int(Error),
		Description: "ERROR: Syntax Error",
	}
}

/*
The filter used by Router

Rulers :

1. if both white list and black list are nil, this filter is all-pass

2. if one of those is not nil, filter as it is

3. if both of those list are not nil, filter as white list
*/
type Filter struct {
	WhiteList  []iosbase.Member
	BlackList  []iosbase.Member
	RejectType []ReqType
	AcceptType []ReqType
}

func (f *Filter) check(req iosbase.Request) bool {
	var memberCheck, typeCheck byte
	if f.WhiteList == nil && f.BlackList == nil {
		memberCheck = byte(0)
	} else if f.WhiteList != nil {
		memberCheck = byte(1)
	} else {
		memberCheck = byte(2)
	}
	if f.AcceptType == nil && f.RejectType == nil {
		typeCheck = byte(0)
	} else if f.AcceptType != nil {
		typeCheck = byte(1)
	} else {
		typeCheck = byte(2)
	}

	var m, t bool

	switch memberCheck {
	case 0:
		m = true
	case 1:
		m = memberContain(req.From, f.WhiteList)
	case 2:
		m = !memberContain(req.From, f.BlackList)
	}

	switch typeCheck {
	case 0:
		t = true
	case 1:
		t = reqTypeContain(req.ReqType, f.AcceptType)
	case 2:
		t = !reqTypeContain(req.ReqType, f.RejectType)
	}

	return m && t
}

func memberContain(a string, c []iosbase.Member) bool {
	for _, m := range c {
		if m.ID == a {
			return true
		}
	}
	return false
}

func reqTypeContain(a int, c []ReqType) bool {
	for _, t := range c {
		if int(t) == a {
			return true
		}
	}
	return false

}
