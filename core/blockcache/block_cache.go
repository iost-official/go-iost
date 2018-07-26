package blockcache

import (
	"bytes"
	"errors"
	"sync"

	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/log"
	//"github.com/iost-official/prototype/log"
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

type BCTType int

const (
	OnCache BCTType = iota
	Singles
)

// BlockCacheTree 缓存链分叉的树结构
type BlockCacheTree struct {
	bc       CachedBlockChain
	children []*BlockCacheTree
	super    *BlockCacheTree
	pool     state.Pool
	bctType  BCTType
}

func newBct(block *block.Block, tree *BlockCacheTree) *BlockCacheTree {
	bct := BlockCacheTree{
		bc:       tree.bc.Copy(),
		children: make([]*BlockCacheTree, 0),
		super:    tree,
		bctType:  Singles,
	}
	bct.bc.Push(block)
	return &bct
}

func (b *BlockCacheTree) popLongest() *BlockCacheTree {
	for _, bct := range b.children {
		if bct.bc.depth == b.bc.depth-1 {
			return bct
		}
	}
	return nil
}

func (b *BlockCacheTree) updateLength() {
	for _, bct := range b.children {
		if bct.bc.parent == &b.bc {
			bct.bc.cachedLength = b.bc.cachedLength + 1
		}
		bct.updateLength()
	}
}

func (b *BlockCacheTree) iterate(fun func(bct *BlockCacheTree) bool) bool {
	if fun(b) {
		return true
	}
	for _, bct := range b.children {
		f := bct.iterate(fun)
		if f {
			return true
		}
	}
	return false
}

var (
	ErrNotFound = errors.New("not found")
	ErrBlock    = errors.New("error block")
	ErrTooOld   = errors.New("block too old")
	ErrDup      = errors.New("block duplicate")
)

type BlockCache interface {
	AddGenesis(block *block.Block) error
	Add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) error

	FindBlockInCache(hash []byte) (*block.Block, error)
	CheckBlock(hash []byte) bool
	LongestChain() block.Chain
	LongestPool() state.Pool
	BlockChain() block.Chain
	BasePool() state.Pool
	SetBasePool(statePool state.Pool) error
	ConfirmedLength() uint64
	BlockConfirmChan() chan uint64
	OnBlockChan() chan *block.Block
	SendOnBlock(blk *block.Block)
}

type BlockCacheImpl struct {
	bc                 block.Chain
	cachedRoot         *BlockCacheTree
	singleBlockRoot    *BlockCacheTree
	hashMap            *sync.Map
	maxDepth           int
	blkConfirmChan     chan uint64
	chConfirmBlockData chan *block.Block
}

func NewBlockCache(chain block.Chain, pool state.Pool, maxDepth int) *BlockCacheImpl {
	h := BlockCacheImpl{
		bc: chain,
		cachedRoot: &BlockCacheTree{
			bc:       NewCBC(chain),
			children: make([]*BlockCacheTree, 0),
			super:    nil,
			pool:     pool,
			bctType:  OnCache,
		},
		singleBlockRoot: &BlockCacheTree{
			bc:       NewCBC(chain),
			children: make([]*BlockCacheTree, 0),
			super:    nil,
			bctType:  Singles,
		},
		hashMap:            new(sync.Map),
		maxDepth:           maxDepth,
		blkConfirmChan:     make(chan uint64, 10),
		chConfirmBlockData: make(chan *block.Block, 100),
	}
	if h.cachedRoot.bc.Top() != nil {
		h.hashMap.Store(string(h.cachedRoot.bc.Top().HeadHash()), h.cachedRoot)
	}
	return &h
}

func (h *BlockCacheImpl) ConfirmedLength() uint64 {
	return h.bc.Length()
}

func (h *BlockCacheImpl) BlockChain() block.Chain {
	return h.bc
}

func (h *BlockCacheImpl) AddGenesis(block *block.Block) error {

	err := h.bc.Push(block)
	if err != nil {
		return err
	}

	err = h.cachedRoot.pool.Flush()
	if err != nil {
		return err
	}
	h.hashMap.Store(string(h.cachedRoot.bc.Top().HeadHash()), h.cachedRoot)
	return nil
}

func (h *BlockCacheImpl) getHashMap(hash []byte) (*BlockCacheTree, bool) {
	rtnI, ok := h.hashMap.Load(string(hash))
	if !ok {
		return nil, false
	}
	rtn, ok := rtnI.(*BlockCacheTree)
	return rtn, ok
}

func (h *BlockCacheImpl) setHashMap(hash []byte, bct *BlockCacheTree) {
	h.hashMap.Store(string(hash), bct)
}

func (h *BlockCacheImpl) Add(blk *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) error {
	var code CacheStatus
	var newTree *BlockCacheTree
	bct, ok := h.getHashMap(blk.HeadHash())
	if ok {
		code, newTree = Duplicate, nil
	} else {
		bct, ok = h.getHashMap(blk.Head.ParentHash)
		if ok {
			code, newTree = h.addSubTree(bct, newBct(blk, bct), verifier)
		} else {
			code, newTree = NotFound, nil
		}
	}

	switch code {
	case Extend:
		fallthrough
	case Fork:
		// Added to cached tree or added to single tree
		h.setHashMap(blk.HeadHash(), newTree)
		if newTree.bctType == OnCache {
			h.addSingles(newTree, verifier)
		} else {
			h.mergeSingles(newTree)
			return ErrNotFound
		}
		h.tryFlush(blk.Head.Version)
	case NotFound:
		// Added as a child of single root
		newTree = newBct(blk, h.singleBlockRoot)
		h.singleBlockRoot.children = append(h.singleBlockRoot.children, newTree)
		h.setHashMap(blk.HeadHash(), newTree)
		h.mergeSingles(newTree)
		return ErrNotFound
	case Duplicate:
		return ErrDup
	case ErrorBlock:
		return ErrBlock
	}
	return nil
}

func (h *BlockCacheImpl) addSingles(newTree *BlockCacheTree, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) {
	block := newTree.bc.Top()
	newChildren := make([]*BlockCacheTree, 0)
	for _, bct := range h.singleBlockRoot.children {
		//fmt.Println(bct.bc.block.Head.ParentHash)
		if bytes.Equal(bct.bc.block.Head.ParentHash, block.HeadHash()) {
			h.addSubTree(newTree, bct, verifier)
		} else {
			newChildren = append(newChildren, bct)
		}
	}
	h.singleBlockRoot.children = newChildren
}

func (h *BlockCacheImpl) mergeSingles(newTree *BlockCacheTree) {
	block := newTree.bc.block
	newChildren := make([]*BlockCacheTree, 0)
	for _, bct := range h.singleBlockRoot.children {
		if bytes.Equal(bct.bc.block.Head.ParentHash, block.HeadHash()) {
			bct.super = newTree
			newTree.children = append(newTree.children, bct)
		} else {
			newChildren = append(newChildren, bct)
		}
	}
	h.singleBlockRoot.children = newChildren
}

func (h *BlockCacheImpl) delSingles() {
	height := h.ConfirmedLength() - 1
	if height%DelSingleBlockTime != 0 {
		return
	}
	newChildren := make([]*BlockCacheTree, 0)
	for _, bct := range h.singleBlockRoot.children {
		if uint64(bct.bc.Top().Head.Number) <= height {
			h.delSubTree(bct)
		} else {
			newChildren = append(newChildren, bct)
		}
	}
	h.singleBlockRoot.children = newChildren
}

func (h *BlockCacheImpl) addSubTree(root *BlockCacheTree, child *BlockCacheTree, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) (CacheStatus, *BlockCacheTree) {
	blk := child.bc.Top()
	newTree := newBct(blk, root)
	if root.bctType == OnCache {
		newPool, err := verifier(blk, root.bc.Top(), root.pool)
		if err != nil {
			h.delSubTree(child)
			log.Log.I("verify block failed. err=%v", err)
			return ErrorBlock, nil
		}
		newTree.pool = newPool
	}
	newTree.bctType = root.bctType
	h.setHashMap(blk.HeadHash(), newTree)
	for _, bct := range child.children {
		h.addSubTree(newTree, bct, verifier)
	}
	root.children = append(root.children, newTree)
	if len(root.children) == 1 {
		return Extend, newTree
	} else {
		return Fork, newTree
	}
}

func (h *BlockCacheImpl) delSubTree(root *BlockCacheTree) {
	h.hashMap.Delete(string(root.bc.Top().HeadHash()))
	for _, bct := range root.children {
		h.delSubTree(bct)
	}
}

func (h *BlockCacheImpl) tryFlush(version int64) {
	for {
		need, newRoot := h.needFlush(version)
		if need {
			for _, bct := range h.cachedRoot.children {
				if bct != newRoot {
					h.delSubTree(bct)
				}
			}
			h.hashMap.Delete(string(h.cachedRoot.bc.Top().HeadHash()))
			h.cachedRoot = newRoot
			h.cachedRoot.bc.Flush()
			err := h.cachedRoot.pool.Flush()
			if err != nil {
				log.Log.E("Database error，failed to tryFlush err:%v", err)
			}
			h.cachedRoot.super = nil
			h.cachedRoot.updateLength()
			h.delSingles()
		} else {
			break
		}
	}
}

func (h *BlockCacheImpl) needFlush(version int64) (bool, *BlockCacheTree) {
	switch version {
	case 0:
		for _, bct := range h.cachedRoot.children {
			if bct.bc.confirmed > h.maxDepth {
				return true, bct
			}
		}
		return false, nil
	case 1:
		if h.cachedRoot.bc.depth > h.maxDepth {
			return true, h.cachedRoot.popLongest()
		}
		return false, nil
	}
	return false, nil
}

func (h *BlockCacheImpl) FindBlockInCache(hash []byte) (*block.Block, error) {
	bct, ok := h.getHashMap(hash)
	if ok {
		return bct.bc.Top(), nil
	}
	return nil, errors.New("block not found")
}

func (h *BlockCacheImpl) CheckBlock(hash []byte) bool {
	if _, err := h.FindBlockInCache(hash); err == nil {
		return true
	}
	if _, err := h.bc.GetBlockByteByHash(hash); err == nil {
		return true
	}
	return false
}

func (h *BlockCacheImpl) LongestChain() block.Chain {
	bct := h.cachedRoot
	for {
		if len(bct.children) == 0 {
			return &bct.bc
		}
		for _, b := range bct.children {
			if b.bc.depth == bct.bc.depth-1 {
				bct = b
				break
			}
		}
	}
}

func (h *BlockCacheImpl) BasePool() state.Pool {
	return h.cachedRoot.pool
}

func (h *BlockCacheImpl) SetBasePool(statePool state.Pool) error {
	h.cachedRoot.pool = statePool
	return nil
}

func (h *BlockCacheImpl) LongestPool() state.Pool {
	bct := h.cachedRoot
	for {
		if len(bct.children) == 0 {
			return bct.pool
		}
		for _, b := range bct.children {
			if b.bc.depth == bct.bc.depth-1 {
				bct = b
				break
			}
		}
	}
}

func (h *BlockCacheImpl) BlockConfirmChan() chan uint64 {
	return h.blkConfirmChan
}

func (h *BlockCacheImpl) OnBlockChan() chan *block.Block {
	return h.chConfirmBlockData
}

func (h *BlockCacheImpl) SendOnBlock(blk *block.Block) {
	h.chConfirmBlockData <- blk
}
