package consensus_common

import (
	"bytes"
	"fmt"

	"errors"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
)

//const (
//	MaxCacheDepth = 6
//)

type CacheStatus int

const (
	Extend CacheStatus = iota
	Fork
	NotFound
	ErrorBlock
)

type BlockCacheTree struct {
	bc       CachedBlockChain
	children []*BlockCacheTree
	super    *BlockCacheTree
	pool     state.Pool
}

func newBct(block *block.Block, tree *BlockCacheTree) *BlockCacheTree {
	bct := BlockCacheTree{
		bc:       tree.bc.Copy(),
		children: make([]*BlockCacheTree, 0),
		super:    tree,
	}

	bct.bc.Push(block)
	return &bct
}

func (b *BlockCacheTree) add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) (CacheStatus, *BlockCacheTree) {
	for _, bct := range b.children {
		code, newTree := bct.add(block, verifier)
		if code != NotFound {
			return code, newTree
		}
	}

	if bytes.Equal(b.bc.Top().Head.Hash(), block.Head.ParentHash) {
		newPool, err := verifier(block, b.bc.Top(), b.pool)
		if err != nil {
			return ErrorBlock, nil
		}

		bct := newBct(block, b)
		bct.pool = newPool
		b.children = append(b.children, bct)
		if len(b.children) == 1 {
			return Extend, bct
		} else {
			return Fork, bct
		}
	}

	return NotFound, nil
}

func (b *BlockCacheTree) findSingles(block *block.Block) (bool, *BlockCacheTree) {
	for _, bct := range b.children {
		found, ret := bct.findSingles(block)
		if found {
			return found, ret
		}
	}
	if b.bc.block != nil && bytes.Equal(b.bc.block.Head.Hash(), block.Head.ParentHash) {
		return true, b
	}
	return false, nil
}

func (b *BlockCacheTree) addSubTree(root *BlockCacheTree, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) {
	block := root.bc.block
	newPool, err := verifier(block, b.bc.Top(), b.pool)
	if err != nil {
		return
	}

	newTree := newBct(block, b)
	newTree.pool = newPool
	b.children = append(b.children, newTree)
	for _, bct := range root.children {
		newTree.addSubTree(bct, verifier)
	}
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
)

type BlockCache interface {
	AddGenesis(block *block.Block) error
	Add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) error
	AddTx(tx *tx.Tx) error
	GetTx() (*tx.Tx, error)
	ResetTxPoool() error

	FindBlockInCache(hash []byte) (*block.Block, error)
	LongestChain() block.Chain
	ConfirmedLength() uint64
}

type BlockCacheImpl struct {
	bc              block.Chain
	cachedRoot      *BlockCacheTree
	singleBlockRoot *BlockCacheTree
	txPool          tx.TxPool
	delTxPool       tx.TxPool
	txPoolCache     tx.TxPool
	maxDepth        int
}

func NewBlockCache(chain block.Chain, pool state.Pool, maxDepth int) *BlockCacheImpl {
	h := BlockCacheImpl{
		bc: chain,
		cachedRoot: &BlockCacheTree{
			bc:       NewCBC(chain),
			children: make([]*BlockCacheTree, 0),
			super:    nil,
			pool:     pool,
		},
		singleBlockRoot: &BlockCacheTree{
			bc: CachedBlockChain{
				block: nil,
			},
			children: make([]*BlockCacheTree, 0),
			super:    nil,
		},
		txPool:    tx.NewTxPoolImpl(),
		delTxPool: tx.NewTxPoolImpl(),
		maxDepth:  maxDepth,
	}
	return &h
}

func (h *BlockCacheImpl) ConfirmedLength() uint64 {
	return h.bc.Length()
}

func (h *BlockCacheImpl) AddGenesis(block *block.Block) error {
	h.bc.Push(block)
	return nil
}

func (h *BlockCacheImpl) Add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) error {
	code, newTree := h.cachedRoot.add(block, verifier)
	switch code {
	case Extend:
		fallthrough
	case Fork:
		// 尝试把single blocks上链
		newChildren := make([]*BlockCacheTree, 0)
		for _, bct := range h.singleBlockRoot.children {
			if bytes.Equal(bct.bc.block.Head.ParentHash, block.Head.Hash()) {
				newTree.addSubTree(bct, verifier)
			} else {
				newChildren = append(newChildren, bct)
			}
		}
		h.singleBlockRoot.children = newChildren
		// 两种情况都可能满足flush
		for {
			// 可能进行多次flush
			need, newRoot := h.needFlush(block.Head.Version)
			if need {
				h.cachedRoot = newRoot
				h.cachedRoot.bc.Flush()
				h.cachedRoot.pool.Flush()
				h.cachedRoot.super = nil
				h.cachedRoot.updateLength()
				for _, tx := range h.cachedRoot.bc.Top().Content {
					h.txPool.Del(tx)
				}
			} else {
				break
			}
		}
	case NotFound:
		// Add to single block tree
		found, bct := h.singleBlockRoot.findSingles(block)
		if !found {
			bct = h.singleBlockRoot
		}
		newTree := &BlockCacheTree{
			bc: CachedBlockChain{
				block: block,
			},
			super:    bct,
			children: make([]*BlockCacheTree, 0),
		}
		bct.children = append(bct.children, newTree)
		return ErrNotFound
	case ErrorBlock:
		return ErrBlock
	}
	return nil
}

func (h *BlockCacheImpl) AddTx(tx *tx.Tx) error {
	//TODO 验证tx是否在blockchain上
	h.txPool.Add(tx)
	return nil
}

func (h *BlockCacheImpl) GetTx() (*tx.Tx, error) {
	for {
		tx := h.txPool.Get()
		h.txPoolCache.Add(tx)
		if ok, _ := h.delTxPool.Has(tx); !ok {
			return tx, nil
		}
	}
}

func (h *BlockCacheImpl) ResetTxPoool() error {
	for h.txPoolCache.Size() > 0 {
		tx := h.txPoolCache.Get()
		h.AddTx(tx)
		h.txPoolCache.Del(tx)
	}
}

func (h *BlockCacheImpl) needFlush(version int64) (bool, *BlockCacheTree) {
	// TODO: 在底层parameter定义的地方定义各种version的const，可以在块生成、验证、此处用
	switch version {
	case 0:
		// DPoS：确认某块的witness数大于maxDepth
		for _, bct := range h.cachedRoot.children {
			if bct.bc.confirmed > h.maxDepth {
				return true, bct
			}
		}
		return false, nil
	case 1:
		// PoW：最长链长度大于maxDepth
		if h.cachedRoot.bc.depth > h.maxDepth {
			return true, h.cachedRoot.popLongest()
		}
		return false, nil
	}
	return false, nil
}

func (h *BlockCacheImpl) FindBlockInCache(hash []byte) (*block.Block, error) {
	var pb *block.Block
	found := h.cachedRoot.iterate(func(bct *BlockCacheTree) bool {
		if bytes.Equal(bct.bc.Top().HeadHash(), hash) {
			pb = bct.bc.Top()
			return true
		} else {
			return false
		}
	})

	if found {
		return pb, nil
	} else {
		return nil, fmt.Errorf("not found")
	}
}

func (h *BlockCacheImpl) LongestChain() block.Chain {
	bct := h.cachedRoot
	h.delTxPool = tx.NewTxPoolImpl()
	for {
		if len(bct.children) == 0 {
			return &bct.bc
		}
		for _, b := range bct.children {
			if b.bc.depth == bct.bc.depth-1 {
				bct = b
				for _, tx := range bct.bc.Top().Content {
					h.delTxPool.Add(tx)
				}
				break
			}
		}
	}
}

func (h *BlockCacheImpl) BasePool() state.Pool {
	return h.cachedRoot.pool
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

//func (h *BlockCacheImpl) FindTx(txHash []byte) (core.Tx, error) {
//	return core.Tx{}, nil
//}

//func (h *BlockCacheImpl) FindTxInCache(txHash []byte) (core.Tx, error) {
//	var tx core.Tx
//	var txp core.TxPoolImpl
//	var err error
//	for _, blk := range h.singleBlocks {
//		txp.Decode(blk.Content)
//		tx, err := txp.Find(txHash)
//		if err == nil {
//			return tx, err
//		}
//	}
//	found := h.cachedRoot.iterate(func(bct *BlockCacheTree) bool {
//		txp.Decode(bct.blk.Content)
//		tx, err = txp.Find(txHash)
//		if err == nil {
//			return true
//		} else {
//			return false
//		}
//	})
//
//	if found {
//		return tx, err
//	} else {
//		return tx, fmt.Errorf("not found")
//	}
//}

//func (h *BlockCacheImpl) LongestPool() core.UTXOPool {
//	bct := h.cachedRoot
//	for {
//		if bct.depth == 0 {
//			return bct.pool
//		}
//		for _, b := range bct.children {
//			if b.depth == bct.depth-1 {
//				bct = b
//				break
//			}
//		}
//	}
//}
