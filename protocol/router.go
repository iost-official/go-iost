package protocol

import (
	"github.com/iost-official/PrototypeWorks/iosbase"
)

type Filter struct {
	WhiteList  []iosbase.Member
	BlackList  []iosbase.Member
	RejectType []ReqType
	AcceptType []ReqType
}

type Router interface {
	FilteredReqChan(filter Filter) (chan iosbase.Request, error)
	Run()
	Stop() error
	SendChan() (chan iosbase.Request, chan iosbase.Response, error)
	ReplyChan() (chan iosbase.Response, error)
}
