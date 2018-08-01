package database

import "github.com/iost-official/Go-IOS-Protocol/db"

type Visitor struct {
	BasicHandler
	MapHandler
	ContractHandler
	MVCCHandler
}

func NewVisitor(cacheLength int, cb *db.MVCCDB) *Visitor {
	db := newChainbaseAdapter(cb)
	cachedDB := NewLRU(cacheLength, db)
	return &Visitor{
		BasicHandler:    BasicHandler{cachedDB},
		MapHandler:      MapHandler{cachedDB},
		ContractHandler: ContractHandler{cachedDB},
		MVCCHandler:     newMVCCHandler(cachedDB, cb),
	}
}
