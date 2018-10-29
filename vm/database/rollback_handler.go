package database

// RollbackHandler rollback delegate
type RollbackHandler struct {
	lru *LRU
	wc  *WriteCache
}

func newRollbackHandler(lru *LRU, wc *WriteCache) RollbackHandler {
	return RollbackHandler{
		lru: lru,
		wc:  wc,
	}
}

// Commit commit a MVCC version
func (m *RollbackHandler) Commit() {
	//ilog.Debug("write cache is:", m.wc.m)
	m.wc.Flush()
	m.wc.Drop()
	//m.lru.Purge()
}

// Rollback rollback to newest MVCC version
func (m *RollbackHandler) Rollback() {
	m.wc.Drop()
	//m.lru.Purge()
}
