package database

// Visitor combine of every handler, to be api of database
type Visitor struct {
	BasicHandler
	MapHandler
	ContractHandler
	BalanceHandler
	CoinHandler
	RollbackHandler
}

// NewVisitor get a visitor of a DB, with cache length determined
func NewVisitor(cacheLength int, cb IMultiValue) *Visitor {
	db := newChainbaseAdapter(cb)
	lruDB := NewLRU(cacheLength, db)
	cachedDB := NewWriteCache(lruDB)
	v := &Visitor{
		BasicHandler:    BasicHandler{cachedDB},
		MapHandler:      MapHandler{cachedDB},
		ContractHandler: ContractHandler{cachedDB},
		CoinHandler:     CoinHandler{cachedDB},
		BalanceHandler:  BalanceHandler{cachedDB},
	}
	v.RollbackHandler = newRollbackHandler(lruDB, cachedDB)
	return v
}
