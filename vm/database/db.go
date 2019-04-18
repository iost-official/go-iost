package database

import "github.com/iost-official/go-iost/core/version"

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
	VoteHandler
}

// NewVisitor get a visitor of a DB, with cache length determined
func NewVisitor(cacheLength int, cb IMultiValue, rules *version.Rules) *Visitor {
	db := newChainbaseAdapter(cb, rules)
	lruDB := NewLRU(cacheLength, db)
	cachedDB := NewWriteCache(lruDB)
	v := &Visitor{
		BasicHandler:    BasicHandler{cachedDB},
		MapHandler:      MapHandler{cachedDB},
		ContractHandler: ContractHandler{cachedDB},
		TokenHandler:    TokenHandler{cachedDB},
		Token721Handler: Token721Handler{cachedDB},
		DelaytxHandler:  DelaytxHandler{cachedDB},
	}
	v.GasHandler = GasHandler{v.BasicHandler, v.MapHandler}
	v.RAMHandler = RAMHandler{v.BasicHandler}
	v.VoteHandler = VoteHandler{v.BasicHandler, v.MapHandler}
	v.RollbackHandler = newRollbackHandler(lruDB, cachedDB)
	return v
}

// NewBatchVisitorRoot get LRU to next step
func NewBatchVisitorRoot(cacheLength int, cb IMultiValue, rules *version.Rules) *LRU {
	db := newChainbaseAdapter(cb, rules)
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
		DelaytxHandler:  DelaytxHandler{cachedDB},
	}
	v.GasHandler = GasHandler{v.BasicHandler, v.MapHandler}
	v.RAMHandler = RAMHandler{v.BasicHandler}
	v.VoteHandler = VoteHandler{v.BasicHandler, v.MapHandler}
	v.RollbackHandler = newRollbackHandler(lruDB, cachedDB)
	return v, watcher
}
