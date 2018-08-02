package database

type Visitor struct {
	BasicHandler
	MapHandler
	ContractHandler
	MVCCHandler
}

func NewVisitor(cacheLength int, cb IMultiValue) *Visitor {
	db := newChainbaseAdapter(cb)
	cachedDB := NewLRU(cacheLength, db)
	return &Visitor{
		BasicHandler:    BasicHandler{cachedDB},
		MapHandler:      MapHandler{cachedDB},
		ContractHandler: ContractHandler{cachedDB},
		MVCCHandler:     newMVCCHandler(cachedDB, cb),
	}
}
