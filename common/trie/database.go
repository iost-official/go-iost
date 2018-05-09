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

	nodes map[common.Hash]*cacheNode
	preimages map[common.Hash][]byte
	seckeybuf [secureKeyLength]byte

	gctime time.Duration // 上次commit后的垃圾回收时间
	gcnodes uint64		 // 上次commit后的垃圾回收node节点数量
	gcsize common.StorageSize

	nodesSize common.StorageSize // node 占用的缓存大小
	preimagesSize common.StorageSize // 镜像占用的缓存大小

	lock sync.RWMutex
}

type cacheNode struct {
	blob []byte
	parents int
	children map[common.Hash]int
}

func NewDatabase(diskdb db.Database) *Database {
	return &Database {
		diskdb: diskdb,
		nodes: map[common.Hash]*cacheNode {
			{}: {children: make(map[common.Hash]int)},
		},
		preimages: make(map[common.Hash][]byte),
	}
}

// DiskDB 返回Trie的一致性数据库索引
func (db *Database) DiskDB() DatabaseReader {
	return db.diskdb
}

func (db *Database) insert(hash common.Hash, blob []byte) {
	if _, ok := db.nodes[hash]; ok {
		return
	}
	db.nodes[hash] = &cacheNode{
		blob: common.CopyBytes
	}
}