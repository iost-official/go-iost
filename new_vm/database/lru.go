package database

import "github.com/hashicorp/golang-lru"

type LRU struct {
	cache *lru.Cache
	db    Database
}

func NewLRU(length int, db Database) *LRU {
	c, err := lru.New(length)
	if err != nil {
		panic(err)
	}
	return &LRU{
		cache: c,
		db:    db,
	}
}

func (m *LRU) Get(key string) (value string) {
	v, ok := m.cache.Get(key)
	if !ok {
		v = m.db.Get(key)
		m.cache.Add(key, v)
	}
	return v.(string)
}
func (m *LRU) Put(key, value string) {
	m.db.Put(key, value)
}
func (m *LRU) Has(key string) bool {
	return m.cache.Contains(key)
}
func (m *LRU) Keys(prefix string) []string {
	return m.db.Keys(prefix)
}
func (m *LRU) Del(key string) {
	m.cache.Remove(key)
	m.db.Del(key)
}
