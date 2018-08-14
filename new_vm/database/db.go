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
	return &Visitor{
		BasicHandler:    BasicHandler{cachedDB},
		MapHandler:      MapHandler{cachedDB},
		ContractHandler: ContractHandler{cachedDB},
		RollbackHandler: newRollbackHandler(cb),
		BalanceHandler:  BalanceHandler{cachedDB},
	}
}
