package database

// RollbackHandler rollback delegate
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

// Commit commit a MVCC version
func (m *RollbackHandler) Commit() {
	m.db.Commit()
}

// Rollback rollback to newest MVCC version
func (m *RollbackHandler) Rollback() {
	m.lru.Purge()
	m.db.Rollback()
}
