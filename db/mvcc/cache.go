package mvcc

import (
	"github.com/iost-official/Go-IOS-Protocol/db/mvcc/map"
	"github.com/iost-official/Go-IOS-Protocol/db/mvcc/trie"
)

type CacheType int

const (
	_ CacheType = iota
	TrieCache
	MapCache
)

type Cache interface {
	Get(key []byte) interface{}
	Put(key []byte, value interface{})
	All(prefix []byte) []interface{}
	Fork() interface{}
	Free()
}

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
