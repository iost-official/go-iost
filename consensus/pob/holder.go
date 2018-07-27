package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
)

type Holder struct {
	self  account.Account
	pool  state.Pool
	spool ServiPool
}

type Watcher struct {
	hp *Holder
}

var Data Holder
