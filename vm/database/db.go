package database

// Visitor combine of every handler, to be api of database
type Visitor struct {
	BasicHandler
	GasHandler
	MapHandler
	ContractHandler
	TokenHandler
	CoinHandler
	RollbackHandler
	DelaytxHandler
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
		TokenHandler:    TokenHandler{cachedDB},
		GasHandler:      GasHandler{cachedDB},
	}
	v.RollbackHandler = newRollbackHandler(lruDB, cachedDB)
	return v
}

// NewBatchVisitorRoot get LRU to next step
func NewBatchVisitorRoot(cacheLength int, cb IMultiValue) *LRU {
	db := newChainbaseAdapter(cb)
	lruDB := NewLRU(cacheLength, db)
	return lruDB
}

// Mapper generator of conflict map
type Mapper interface {
	Map() map[string]Access
}

// NewBatchVisitor get visitor with mapper
func NewBatchVisitor(lruDB *LRU) (*Visitor, Mapper) {
	cachedDB := NewWriteCache(lruDB)

	watcher := NewWatcher(cachedDB)
	v := &Visitor{
		BasicHandler:    BasicHandler{watcher},
		MapHandler:      MapHandler{watcher},
		ContractHandler: ContractHandler{watcher},
		CoinHandler:     CoinHandler{watcher},
		TokenHandler:    TokenHandler{watcher},
		GasHandler:      GasHandler{watcher},
	}
	v.RollbackHandler = newRollbackHandler(lruDB, cachedDB)
	return v, watcher
}
