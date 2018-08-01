package database

import "github.com/iost-official/Go-IOS-Protocol/chainbase"

type Visitor struct {
	BasicHandler
	MapHandler
	ContractHandler
}

func NewVisitor(cacheLength int, cb *chainbase.Chainbase) *Visitor {
	db := newChainbaseAdapter(cb)
	cachedDB := NewLRU(cacheLength, db)
	return &Visitor{
		BasicHandler:    BasicHandler{cachedDB},
		MapHandler:      MapHandler{cachedDB},
		ContractHandler: ContractHandler{cachedDB},
	}
}
