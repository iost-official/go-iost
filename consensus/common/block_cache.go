package consensus_common

import (
	"bytes"
	"fmt"
	"github.com/iost-official/prototype/core"
	"github.com/iost-official/prototype/verifier"
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

func newBct(block *core.Block, tree *BlockCacheTree) *BlockCacheTree {
	bct := BlockCacheTree{
		depth:    0,
		bc:       tree.bc.Copy(),
		children: make([]*BlockCacheTree, 0),
		super:    tree,
	}

	bct.bc.Push(block)
	return &bct
}

func (b *BlockCacheTree) add(block *core.Block, verifier verifier.BlockVerifier ) CacheStatus {

	var code CacheStatus
	for _, bct := range b.children {
		code = bct.add(block, verifier)
		if code == Extend {
			if bct.depth == b.depth {
				b.depth++
				return Extend
			} else {
				return Fork
			}
		} else if code == Fork {
			return Fork
		} else if code == ErrorBlock {
			return ErrorBlock
		}
	}

	if bytes.Equal(b.bc.Top().Head.Hash(), block.Head.ParentHash) {
		if !verifier(block, &b.bc) {
			return ErrorBlock
		}

		if len(b.children) == 0 {
			b.children = append(b.children, newBct(block, b))
			b.depth++
			return Extend
		} else {
			return Fork
		}
	}

	return NotFound
}

func (b *BlockCacheTree) pop() *BlockCacheTree {

	for _, bct := range b.children {
		if bct.depth == b.depth-1 {
			return bct
		}
	}
	return nil
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
	Add(block *core.Block, verifier func(blk *core.Block, chain core.BlockChain) bool) error
	FindBlockInCache(hash []byte) (*core.Block, error)
	LongestChain() core.BlockChain
}

type BlockCacheImpl struct {
	bc           core.BlockChain
	cachedRoot   *BlockCacheTree
	singleBlocks []*core.Block
	maxDepth     int
}

func NewBlockCache(chain core.BlockChain, maxDepth int) *BlockCacheImpl {
	h := BlockCacheImpl{
		bc: chain,
		cachedRoot: &BlockCacheTree{
			depth:    0,
			bc:       NewCBC(chain),
			children: make([]*BlockCacheTree, 0),
			super:    nil,
		},
		singleBlocks: make([]*core.Block, 0),
		maxDepth:     maxDepth,
	}
	return &h
}

func (h *BlockCacheImpl) Add(block *core.Block, verifier func(blk *core.Block, chain core.BlockChain) bool) error {
	code := h.cachedRoot.add(block, verifier)
	switch code {
	case Extend:
		if h.cachedRoot.depth > h.maxDepth {
			h.cachedRoot = h.cachedRoot.pop()
			h.cachedRoot.super = nil
			h.cachedRoot.bc.Flush()
		}
		fallthrough
	case Fork:
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

func (h *BlockCacheImpl) FindBlockInCache(hash []byte) (*core.Block, error) {
	var pb *core.Block
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

func (h *BlockCacheImpl) LongestChain() core.BlockChain {
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
