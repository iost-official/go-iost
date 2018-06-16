package trie

import (
	"sync"
	"time"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/db"
)

var secureKeyPrefix = []byte("secure-key-")

const secureKeyLength = 11 + 32

type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)

	Has(key []byte) (bool, error)
}

type Database struct {
	diskdb db.Database

	nodes     map[common.Hash]*cachedNode
	preimages map[common.Hash][]byte
	seckeybuf [secureKeyLength]byte

	gctime  time.Duration
	gcnodes uint64
	gcsize  common.StorageSize

	nodesSize     common.StorageSize
	preimagesSize common.StorageSize

	lock sync.RWMutex
}

type cachedNode struct {
	blob     []byte
	parents  int
	children map[common.Hash]int
}

func NewDatabase(diskdb db.Database) *Database {
	return &Database{
		diskdb: diskdb,
		nodes: map[common.Hash]*cachedNode{
			{}: {children: make(map[common.Hash]int)},
		},
		preimages: make(map[common.Hash][]byte),
	}
}

func (db *Database) DiskDB() DatabaseReader {
	return db.diskdb
}

func (db *Database) Insert(hash common.Hash, blob []byte) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.insert(hash, blob)
}

func (db *Database) insert(hash common.Hash, blob []byte) {
	if _, ok := db.nodes[hash]; ok {
		return
	}
	db.nodes[hash] = &cachedNode{
		blob:     common.CopyBytes(blob),
		children: make(map[common.Hash]int),
	}
	db.nodesSize += common.StorageSize(common.HashLength + len(blob))
}

func (db *Database) insertPreimage(hash common.Hash, preimage []byte) {
	if _, ok := db.preimages[hash]; ok {
		return
	}
	db.preimages[hash] = common.CopyBytes(preimage)
	db.preimagesSize += common.StorageSize(common.HashLength + len(preimage))
}

func (db *Database) Node(hash common.Hash) ([]byte, error) {
	db.lock.RLock()
	node := db.nodes[hash]
	db.lock.RUnlock()

	if node != nil {
		return node.blob, nil
	}
	return db.diskdb.Get(hash[:])
}

func (db *Database) preimage(hash common.Hash) ([]byte, error) {
	db.lock.RLock()
	preimage := db.preimages[hash]
	db.lock.RUnlock()

	if preimage != nil {
		return preimage, nil
	}
	return db.diskdb.Get(db.secureKey(hash[:]))
}

func (db *Database) secureKey(key []byte) []byte {
	buf := append(db.seckeybuf[:0], secureKeyPrefix...)
	buf = append(buf, key...)
	return buf
}

func (db *Database) Nodes() []common.Hash {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var hashes = make([]common.Hash, 0, len(db.nodes))
	for hash := range db.nodes {
		if hash != (common.Hash{}) {
			hashes = append(hashes, hash)
		}
	}
	return hashes
}

func (db *Database) Reference(child common.Hash, parent common.Hash) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	db.reference(child, parent)
}

func (db *Database) reference(child common.Hash, parent common.Hash) {
	node, ok := db.nodes[child]
	if !ok {
		return
	}
	if _, ok = db.nodes[parent].children[child]; ok && parent != (common.Hash{}) {
		return
	}
	node.parents++
	db.nodes[parent].children[child]++
}

func (db *Database) Dereference(child common.Hash, parent common.Hash) {
	db.lock.Lock()
	defer db.lock.Unlock()

	nodes, storage, start := len(db.nodes), db.nodesSize, time.Now()
	db.dereference(child, parent)

	db.gcnodes += uint64(nodes - len(db.nodes))
	db.gcsize += storage - db.nodesSize
	db.gctime += time.Since(start)
}

func (db *Database) dereference(child common.Hash, parent common.Hash) {
	node := db.nodes[parent]

	node.children[child]--
	if node.children[child] == 0 {
		delete(node.children, child)
	}
	node, ok := db.nodes[child]
	if !ok {
		return
	}
	node.parents--
	if node.parents == 0 {
		for hash := range node.children {
			db.dereference(hash, child)
		}
		delete(db.nodes, child)
		db.nodesSize -= common.StorageSize(common.HashLength + len(node.blob))
	}
}

func (db *Database) Commit(node common.Hash, report bool) error {
	db.lock.RLock()

	batch := db.diskdb

	for hash, preimage := range db.preimages {
		if err := batch.Put(db.secureKey(hash[:]), preimage); err != nil {
			db.lock.RUnlock()
			return err
		}
	}
	if err := db.commit(node, batch); err != nil {
		db.lock.RUnlock()
		return err
	}
	db.lock.RUnlock()

	db.lock.Lock()
	defer db.lock.Unlock()

	db.preimages = make(map[common.Hash][]byte)
	db.preimagesSize = 0

	db.uncache(node)

	db.gcnodes, db.gcsize, db.gctime = 0, 0, 0

	return nil
}

func (db *Database) commit(hash common.Hash, batch db.Database) error {
	node, ok := db.nodes[hash]
	if !ok {
		return nil
	}
	for child := range node.children {
		if err := db.commit(child, batch); err != nil {
			return err
		}
	}
	if err := batch.Put(hash[:], node.blob); err != nil {
		return err
	}
	return nil
}

func (db *Database) uncache(hash common.Hash) {
	node, ok := db.nodes[hash]
	if !ok {
		return
	}
	for child := range node.children {
		db.uncache(child)
	}
	delete(db.nodes, hash)
	db.nodesSize -= common.StorageSize(common.HashLength + len(node.blob))
}

func (db *Database) Size() common.StorageSize {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return db.nodesSize + db.preimagesSize
}
