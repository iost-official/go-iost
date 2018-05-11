package trie

import (
	"errors"
	"github.com/iost-official/prototype/common"
)

var ErrNotRequested = errors.New("not requested")

var ErrAlreadyProcessed = errors.New("already processed")

type request struct {
	hash common.Hash
	data []byte
	raw bool

	parents []*request
	depth int
	deps int

	callback LeafCallback
}

type SyncResult struct {
	Hash common.Hash
	Data []byte
}

type syncMemBatch struct {
	batch map[common.Hash][]byte
	order []common.Hash
}

func newSyncMemBatch() *syncMemBatch {
	return &syncMemBatch{
		batch: make(map[common.Hash][]byte),
		order: make([]common.Hash, 0, 256),
	}
}

type TrieSync struct {
	database DatabaseReader
	membatch *syncMemBatch
	requests map[common.Hash]*request
	queue *prque.Prque
}