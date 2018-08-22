package database

import "github.com/hashicorp/golang-lru"

// LRU ...
type LRU struct {
	cache *lru.Cache
	db    database
}

// NewLRU ...
func NewLRU(length int, db database) *LRU {
	if length <= 0 {
		return &LRU{
			cache: nil,
			db:    db,
		}
	}

	c, err := lru.New(length)
	if err != nil {
		panic(err)
	}
	return &LRU{
		cache: c,
		db:    db,
	}
}

// Get ...
func (m *LRU) Get(key string) (value string) {
	if m.cache == nil {
		return m.db.Get(key)
	}
	v, ok := m.cache.Get(key)
	if !ok {
		v = m.db.Get(key)
		m.cache.Add(key, v)
	}
	return v.(string)
}

// Put ...
func (m *LRU) Put(key, value string) {
	if m.cache != nil && m.cache.Contains(key) {
		m.cache.Add(key, value)
	}
	m.db.Put(key, value)
}

// Has ...
func (m *LRU) Has(key string) bool {
	if m.cache == nil {
		return m.db.Has(key)
	}
	return m.cache.Contains(key)
}

// Keys ...
func (m *LRU) Keys(prefix string) []string {
	return m.db.Keys(prefix)
}

// Del ...
func (m *LRU) Del(key string) {
	if m.cache != nil {
		m.cache.Remove(key)
	}
	m.db.Del(key)
}

// Purge ...
func (m *LRU) Purge() {
	if m.cache != nil {
		m.cache.Purge()
	}
}
