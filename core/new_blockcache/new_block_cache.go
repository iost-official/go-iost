package blockcache

import (
	"errors"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	blockCachedLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "block_cached_length",
			Help: "Length of cached block chain",
		},
	)
)

func init() {
	prometheus.MustRegister(blockCachedLength)
}

type CacheStatus int

const (
	Extend CacheStatus = iota
	Fork
	NotFound
	ErrorBlock
	Duplicate
)

const (
	DelSingleBlockTime uint64 = 10
)

type BCNType int

const (
	Linked BCNType = iota
	Single
)

type BlockCacheNode struct {
	Block                 *block.Block
	commit                string
	Parent                *BlockCacheNode
	Children              []*BlockCacheNode
	Type                  BCNType
	Number                uint64
	Witness               string
	ConfirmUntil          uint64
	LastWitnessListNumber uint64
	PendingWitnessList    []*string
	Extension             []byte
}

func NewBCN(parent *BlockCacheNode, block *block.Block, nodeType BCNType) *BlockCacheNode {
	return nil
}

type BlockCache struct {
	linkedTree *BlockCacheNode
	singleTree *BlockCacheNode
	Head       *BlockCacheNode
	hash2node  *sync.Map
}

var (
	ErrNotFound = errors.New("not found")
	ErrBlock    = errors.New("error block")
	ErrTooOld   = errors.New("block too old")
	ErrDup      = errors.New("block duplicate")
)

func NewBlockCache() *BlockCache {
	return nil
}

func (bc *BlockCache) Add(blk *block.Block) (*BlockCacheNode, error) {
	return nil, nil
}

func (bc *BlockCache) Flush(node *BlockCacheNode) {
	return
}

func (bc *BlockCache) Find(blkHash []byte) *BlockCacheNode {
	return nil
}
