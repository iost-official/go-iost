package trie

import (
	"github.com/iost-official/prototype/common"
	"errors"
)

type Iterator struct {
	nodeIt NodeIterator

	Key []byte
	Value []byte
	Err error
}

// Iterator is a key-value trie iterator that traverses a Trie.
func NewIterator(it NodeIterator) *Iterator {
	return &Iterator{
		nodeIt: it,
	}
}

func (it *Iterator) Next() bool {
	for it.nodeIt.Next(true) {
		if it.nodeIt.Leaf() {
			it.Key = it.nodeIt.LeafKey()
			it.Value = it.nodeIt.LeafBlob()
			return true
		}
	}
	it.Key = nil
	it.Value = nil
	it.Err = it.nodeIt.Error()
	return false
}

// pre-order
type NodeIterator interface {
	Next(bool) bool
	Error() error
	Hash() common.Hash
	Parent() common.Hash
	Path() []byte

	Leaf() bool
	LeafBlob() []byte
	LeafKey() []byte
}

type nodeIteratorState struct {
	hash common.Hash
	node node
	parent common.Hash
	index int
	pathlen int
}

type nodeIterator struct {
	trie *Trie
	stack []*nodeIteratorState
	path []byte
	err error
}

var iteratorEnd = errors.New("end of iteration")

type seekError struct {
	key []byte
	err error
}

func (e seekError) Error() string {
	return "seek error: " + e.err.Error()
}

func newNodeIterator(trie *Trie, start []byte) NodeIterator {
	if trie.Hash() == emptyState {
		return new(nodeIteratorState)
	}
}