package database

import "github.com/iost-official/Go-IOS-Protocol/db"

type MVCCHandler struct {
	cache      *LRU
	db         *db.MVCCDB
	currentTag string
}

func newMVCCHandler(cache *LRU, db *db.MVCCDB) MVCCHandler {
	return MVCCHandler{
		cache:      cache,
		db:         db,
		currentTag: "",
	}
}

func (m *MVCCHandler) Checkout(tag string) {
	if tag == m.currentTag {
		return
	} else {
		m.cache.Purge()
		//m.db.
	}
}
