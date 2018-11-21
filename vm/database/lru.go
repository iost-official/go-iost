package database

import (
	"github.com/hashicorp/golang-lru"
)

// LRU lru cache
type LRU struct {
	cache *lru.Cache
	db    database
}

// NewLRU make a new lru
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

// Get get from cache
func (m *LRU) Get(key string) (value string) {
	if m.cache == nil {
		return m.db.Get(key)
	}
	v, ok := m.cache.Get(key)
	if !ok {
		value = m.db.Get(key)
		if value != "" && value != "n" {
			m.cache.Add(key, value)
		}
		return value
	}
	return v.(string)
}

// Put put kv to cache
func (m *LRU) Put(key, value string) {
	if m.cache != nil && m.cache.Contains(key) {
		m.cache.Add(key, value)
	}
	m.db.Put(key, value)
}

// Has if key exist
func (m *LRU) Has(key string) bool {
	if m.cache == nil {
		return m.db.Has(key)
	}
	ok := m.cache.Contains(key)
	if !ok {
		ok = m.db.Has(key)
		if ok {
			v := m.db.Get(key)
			m.cache.Add(key, v)
		}
	}
	return ok
}

// Keys list keys under prefix, do nothing
//func (m *LRU) Keys(prefix string) []string {
//	return m.db.Keys(prefix)
//}

// Del delete key from cache
func (m *LRU) Del(key string) {
	if m.cache != nil {
		m.cache.Remove(key)
	}
	m.db.Del(key)
}

// Purge delete all keys
func (m *LRU) Purge() {
	if m.cache != nil {
		m.cache.Purge()
	}
}
