package trie

import (
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/common"
	"time"
	"sync"
)

// 数据库关键字前缀，用来存储node信息
var secureKeyPrefix = []byte("secure-key-")

// 关键字长度是前缀长度加上32位哈希值
const secureKeyLength = 11 + 32

type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)
	Has(key []byte) (bool, error)
}

// Database 是trie数据结构和硬盘数据库的中间层
// 利用内存加速trie写速度，仅周期性地向硬盘写
type Database struct {
	diskdb db.Database

	nodes map[common.Hash]*cachedNode
	preimages map[common.Hash][]byte
	seckeybuf [secureKeyLength]byte

	gctime time.Duration // 上次commit后的垃圾回收时间
	gcnodes uint64		 // 上次commit后的垃圾回收node节点数量
	gcsize common.StorageSize

	nodesSize common.StorageSize // node 占用的缓存大小
	preimagesSize common.StorageSize // 镜像占用的缓存大小

	lock sync.RWMutex
}

type cachedNode struct {
	blob []byte
	parents int
	children map[common.Hash]int
}

func NewDatabase(diskdb db.Database) *Database {
	return &Database {
		diskdb: diskdb,
		nodes: map[common.Hash]*cachedNode {
			{}: {children: make(map[common.Hash]int)},
		},
		preimages: make(map[common.Hash][]byte),
	}
}

// DiskDB 返回Trie的一致性数据库索引
func (db *Database) DiskDB() DatabaseReader {
	return db.diskdb
}


// 将节点插入内存数据库中
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

// 从数据库中取回一个节点，如果节点不在缓存中，则从数据库中返回
func (db *Database) Node(hash common.Hash) ([]byte, error) {
	// 优先从内存中返回节点信息
	db.lock.RLock()
	node := db.nodes[hash]
	db.lock.RUnlock()

	if node != nil {
		return node.blob, nil
	}
	// 内存中不存在，返回数据库中的结果
	return db.diskdb.Get(hash[:])
}

// 从数据库中取回一个节点镜像，优先从内存中取，如果内存中不存在，则返回数据库中的结果
func (db *Database) preimage(hash common.Hash) ([]byte, error) {
	// 优先从内存中读取结果
	db.lock.RLock()
	preimage := db.preimages[hash]
	db.lock.RUnlock()

	if preimage != nil {
		return preimage, nil
	}
	// 内存中不存在，返回数据库中的结果
	return db.diskdb.Get(db.secureKey(hash[:]))
}

// 调用函数需要自己拷贝结果，因为该函数下次被调用时，之前的结果会被覆盖
func (db *Database) secureKey(key []byte) []byte {
	buf := append(db.seckeybuf[:0], secureKeyPrefix...)
	buf = append(buf, key...)
	return buf
}

// 返回缓存中所有节点的哈希值
// 函数时间消耗很高，慎用
func (db *Database) Nodes() []common.Hash {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var hashes = make([]common.Hash, 0, len(db.nodes))
	for hash := range db.nodes {
		if hash != (common.Hash{}) { // 特判根节点
			hashes = append(hashes, hash)
		}
	}
	return hashes
}

// 添加一个从父节点指向子节点的引用
func (db *Database) Reference(child common.Hash, parent common.Hash) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	db.reference(child, parent)
}

func (db *Database) reference(child common.Hash, parent common.Hash) {
	// 数据库中的节点，直接跳过
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

// 删除从根节点指向子节点的指针
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
	// 如果子节点的引用计数为0，则递归删除子节点
	node.parents--
	if node.parents == 0 {
		for hash := range node.children {
			db.dereference(hash, child)
		}
		delete(db.nodes, child)
		db.nodesSize -= common.StorageSize(common.HashLength + len(node.blob))
	}
}

// 暴力遍历一个点的所有子节点，将其写回硬盘
func (db *Database) Commit(node common.Hash, report bool) error {
	db.lock.RLock()

	batch := db.diskdb

	for hash, preimage := range db.preimages {
		if err := batch.Put(db.secureKey(hash[:]), preimage); err != nil {
			db.lock.RUnlock()
			return err
		}
	}
	db.lock.RUnlock()

	// Write successful, clear out the flushed data
	db.lock.Lock()
	defer db.lock.Unlock()

	db.preimages = make(map[common.Hash][]byte)
	db.preimagesSize = 0

	db.uncache(node)

	// commit 之后重置垃圾回收参数
	db.gcnodes, db.gcsize, db.gctime = 0, 0, 0

	return nil
}

// TODO: 用 数据库中的batch接口加速写操作，缓存一部分写硬盘操作，每隔一定周期再写
func (db *Database) commit(hash common.Hash, batch db.Database) error {
	// 如果节点不存在，则返回上一个commit版本的值
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

// 返回当前缓存占用空间的大小
func (db *Database) Size() common.StorageSize {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return db.nodesSize + db.preimagesSize
}