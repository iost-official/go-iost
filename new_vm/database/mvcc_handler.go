package database

type MVCCHandler struct {
	cache      *LRU
	db         IMultiValue
	currentTag string
}

func newMVCCHandler(cache *LRU, db IMultiValue) MVCCHandler {
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
		//m.db.Checkout(tag) // 注意这个
		m.currentTag = tag
	}
}
