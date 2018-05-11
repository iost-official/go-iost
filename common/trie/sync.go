package trie

import (
	"errors"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/common/trie/prque"
	"fmt"
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

func NewTrieSync(root common.Hash, database DatabaseReader, callback LeafCallback) *TrieSync {
	ts := &TrieSync {
		database: database,
		membatch: newSyncMemBatch(),
		requests: make(map[common.Hash]*request),
		queue: prque.New(),
	}
	ts.AddSubTrie(root, 0, common.Hash{}, callback)
	return ts
}

func (s *TrieSync) AddSubTrie(root common.Hash, depth int, parent common.Hash, callback LeafCallback) {
	if root == emptyRoot {
		return
	}
	if _, ok := s.membatch.batch[root]; ok {
		return
	}
	key := root.Bytes()
	blob, _ := s.database.Get(key)
	if local, err := decodeNode(key, blob, 0); local != nil && err == nil {
		return
	}
	req := &request {
		hash: root,
		depth: depth,
		callback: callback,
	}
	if parent != (common.Hash{}) {
		ancestor := s.requests[parent]
		if ancestor == nil {
			panic(fmt.Sprintf("sub-trie ancestor not found: %x", parent))
		}
		ancestor.deps++
		req.parents = append(req.parents, ancestor)
	}
	s.schedule(req)
}

func (s *TrieSync) AddRawEntry(hash common.Hash, depth int, parent common.Hash) {
	if hash == emptyState {
		return
	}
	if _, ok := s.membatch.batch[hash]; ok {
		return
	}
	if ok, _ := s.database.Has(hash.Bytes()); ok {
		return
	}
	req := &request{
		hash: hash,
		raw: true,
		depth: depth,
	}
	if parent != (common.Hash{}) {
		ancestor := s.requests[parent]
		if ancestor == nil {
			panic(fmt.Sprintf("raw-entry ancestor not found: %x", parent))
		}
		ancestor.deps++
		req.parents = append(req.parents, ancestor)
	}
	s.schedule(req)
}

func (s *TrieSync) Missing(max int) []common.Hash {
	requests := []common.Hash{}
	for !s.queue.Empty() && (max == 0 || len(requests) < max) {
		requests = append(requests, s.queue.PopItem().(common.Hash))
	}
	return requests
}

func (s *TrieSync) schedule(req *request) {
	if old, ok := s.requests[req.hash]; ok {
		old.parents = append(old.parents, req.parents...)
		return
	}
	s.queue.Push(req.hash, float32(req.depth))
	s.requests[req.hash] = req
}
