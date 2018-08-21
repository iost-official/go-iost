package database

type Visitor struct {
	BasicHandler
	MapHandler
	ContractHandler
	BalanceHandler
	RollbackHandler
}

func NewVisitor(cacheLength int, cb IMultiValue) *Visitor {
	db := newChainbaseAdapter(cb)
	cachedDB := NewLRU(cacheLength, db)
	v := &Visitor{
		BasicHandler:    BasicHandler{cachedDB},
		MapHandler:      MapHandler{cachedDB},
		ContractHandler: ContractHandler{cachedDB},
		BalanceHandler:  BalanceHandler{cachedDB},
	}
	v.RollbackHandler = newRollbackHandler(cb, cachedDB)
	return v
}
