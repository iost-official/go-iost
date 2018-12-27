package database

// Visitor combine of every handler, to be api of database
type Visitor struct {
	BasicHandler
	MapHandler
	ContractHandler
	TokenHandler
	Token721Handler
	RollbackHandler
	DelaytxHandler
	GasHandler
	RAMHandler
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
		TokenHandler:    TokenHandler{cachedDB},
		Token721Handler: Token721Handler{cachedDB},
		RAMHandler:      RAMHandler{cachedDB},
		DelaytxHandler:  DelaytxHandler{cachedDB},
	}
	v.GasHandler = GasHandler{v.BasicHandler, v.MapHandler}
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
		TokenHandler:    TokenHandler{watcher},
		Token721Handler: Token721Handler{watcher},
		RAMHandler:      RAMHandler{watcher},
		DelaytxHandler:  DelaytxHandler{cachedDB},
	}
	v.GasHandler = GasHandler{v.BasicHandler, v.MapHandler}
	v.RollbackHandler = newRollbackHandler(lruDB, cachedDB)
	return v, watcher
}
