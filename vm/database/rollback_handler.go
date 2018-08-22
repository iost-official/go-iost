package database

// RollbackHandler ...
type RollbackHandler struct {
	db  IMultiValue
	lru *LRU
}

func newRollbackHandler(db IMultiValue, lru *LRU) RollbackHandler {
	return RollbackHandler{
		db:  db,
		lru: lru,
	}
}

// Commit ...
func (m *RollbackHandler) Commit() {
	m.db.Commit()
}

// Rollback ...
func (m *RollbackHandler) Rollback() {
	m.lru.Purge()
	m.db.Rollback()
}
