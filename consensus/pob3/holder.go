package pob3

import (
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/state"
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
