package consensus_common

import (
	"bytes"
	"fmt"

	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
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
	depth    int
	bc       CachedBlockChain
	children []*BlockCacheTree
	super    *BlockCacheTree
}

func newBct(block *block.Block, tree *BlockCacheTree) *BlockCacheTree {
	bct := BlockCacheTree{
		depth:    0,
		bc:       tree.bc.Copy(),
		children: make([]*BlockCacheTree, 0),
		super:    tree,
	}

	bct.bc.Push(block)
	return &bct
}

func (b *BlockCacheTree) add(block *block.Block, verifier func(blk *block.Block, chain block.Chain) (bool, state.Pool)) CacheStatus {
	var code CacheStatus
	for _, bct := range b.children {
		code = bct.add(block, verifier)
		if code == Extend {
			if bct.depth == b.depth {
				b.depth++
			}
			return Extend
		} else if code == Fork {
			return Fork
		} else if code == ErrorBlock {
			return ErrorBlock
		}
	}

	if bytes.Equal(b.bc.Top().Head.Hash(), block.Head.ParentHash) {
		result, newPool := verifier(block, &b.bc)
		if !result {
			return ErrorBlock
		}

		bct := newBct(block, b)
		bct.bc.SetStatePool(newPool)
		b.children = append(b.children, bct)
		if len(b.children) == 1 {
			b.depth++
			return Extend
		} else {
			return Fork
		}
	}

	return NotFound
}

func (b *BlockCacheTree) popLongest() *BlockCacheTree {
	for _, bct := range b.children {
		if bct.depth == b.depth-1 {
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

type BlockCache interface {
	Add(block *block.Block, verifier func(blk *block.Block, chain block.Chain) (bool, state.Pool)) error
	FindBlockInCache(hash []byte) (*block.Block, error)
	LongestChain() block.Chain
}

type BlockCacheImpl struct {
	bc           block.Chain
	cachedRoot   *BlockCacheTree
	singleBlocks []*block.Block
	maxDepth     int
}

func NewBlockCache(chain block.Chain, maxDepth int) *BlockCacheImpl {
	h := BlockCacheImpl{
		bc: chain,
		cachedRoot: &BlockCacheTree{
			depth:    0,
			bc:       NewCBC(chain),
			children: make([]*BlockCacheTree, 0),
			super:    nil,
		},
		singleBlocks: make([]*block.Block, 0),
		maxDepth:     maxDepth,
	}
	return &h
}

func (h *BlockCacheImpl) Add(block *block.Block, verifier func(blk *block.Block, chain block.Chain) (bool, state.Pool)) error {
	code := h.cachedRoot.add(block, verifier)
	switch code {
	case Extend:
		fallthrough
	case Fork:
		// 两种情况都可能满足flush
		for {
			// 可能进行多次flush
			need, newRoot := h.needFlush(block.Head.Version)
			if need {
				h.cachedRoot = newRoot
				h.cachedRoot.bc.Flush()
				h.cachedRoot.super = nil
				h.cachedRoot.updateLength()
			} else {
				break
			}
		}
		// TODO 考虑递归Add情况singleBlocks很多冗余
		for _, blk := range h.singleBlocks {
			h.Add(blk, verifier)
		}
	case NotFound:
		h.singleBlocks = append(h.singleBlocks, block)
	case ErrorBlock:
		return fmt.Errorf("error found")
	}
	return nil
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
		if h.cachedRoot.depth > h.maxDepth {
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
	if bct.depth == 0 {
		return &h.cachedRoot.bc
	}
	for {
		if bct.depth == 0 {
			return &bct.bc
		}
		for _, b := range bct.children {
			if b.depth == bct.depth-1 {
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
