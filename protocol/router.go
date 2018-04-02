package protocol

import (
	"github.com/iost-official/PrototypeWorks/iosbase"
)

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


type Filter struct {
	WhiteList  []iosbase.Member
	BlackList  []iosbase.Member
	RejectType []ReqType
	AcceptType []ReqType
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
		Description: "Accepted",
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

func syntaxError(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Error),
		Description: "ERROR: Syntax Error",
	}
}


type Router interface {
	FilteredInChan(filter Filter) (chan iosbase.Request, error)
	FilteredOutChan(filter Filter) (chan iosbase.Request, error)
	Run()
	Stop() error
	SendChan() (chan iosbase.Request, chan iosbase.Response, error)
	ReplyChan() (chan iosbase.Response, error)
}
