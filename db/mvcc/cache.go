package mvcc

import (
	mvccmap "github.com/iost-official/go-iost/v3/db/mvcc/map"
	"github.com/iost-official/go-iost/v3/db/mvcc/trie"
)

// CacheType is the cache type
type CacheType int

// The cache type constant
const (
	_ CacheType = iota
	TrieCache
	MapCache
)

// Cache is the cache interface
type Cache interface {
	Get(key []byte) any
	Put(key []byte, value any)
	All(prefix []byte) []any
	Fork() any
	Free()
}

// NewCache returns the specify type cache
func NewCache(t CacheType) Cache {
	switch t {
	case TrieCache:
		return trie.New()
	case MapCache:
		return mvccmap.New()
	default:
		return trie.New()
	}
}
