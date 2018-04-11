package pow

import (
	"bytes"
	"fmt"
	"github.com/iost-official/prototype/core"
)

const (
	MaxCacheDepth = 6
)

type BlockCacheTree struct {
	depth    int
	blk      *core.Block
	pool     core.UTXOPool
	children []*BlockCacheTree
	super    *BlockCacheTree
}

type CacheStatus int

const (
	Extend CacheStatus = iota
	Fork
	NotFound
	ErrorBlock
)

func newBct(block *core.Block, tree *BlockCacheTree) *BlockCacheTree {
	pool := tree.pool.Copy()
	pool.Transact(block)
	bct := BlockCacheTree{
		depth:    0,
		blk:      block,
		pool:     pool,
		children: make([]*BlockCacheTree, 0),
		super:    tree,
	}
	return &bct
}

func (b *BlockCacheTree) add(block *core.Block) CacheStatus {

	var code CacheStatus
	for _, bct := range b.children {
		code = bct.add(block)
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

	if bytes.Equal(b.blk.Head.Hash(), block.Head.SuperHash) {
		if IsLegalBlock(block, b.pool) != nil {
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

type Holder struct {
	bc           core.BlockChain
	pool         core.UTXOPool
	cachedRoot   *BlockCacheTree
	singleBlocks []*core.Block
}

func NewHolder(chain core.BlockChain, pool core.UTXOPool) Holder {
	h := Holder{
		bc: chain,
		cachedRoot: &BlockCacheTree{
			depth:    0,
			blk:      chain.Top(),
			children: make([]*BlockCacheTree, 0),
			super:    nil,
		},
		pool:         pool,
		singleBlocks: make([]*core.Block, 0),
	}

	h.cachedRoot.pool = h.pool.Copy()
	return h
}

func (h *Holder) Add(block *core.Block) error {
	code := h.cachedRoot.add(block)
	switch code {
	case Extend:
		if h.cachedRoot.depth > MaxCacheDepth {
			h.cachedRoot = h.cachedRoot.pop()
			h.cachedRoot.super = nil
			h.bc.Push(h.cachedRoot.blk)
		}
		fallthrough
	case Fork:
		for _, blk := range h.singleBlocks {
			h.Add(blk)
		}
	case NotFound:
		h.singleBlocks = append(h.singleBlocks, block)
	case ErrorBlock:
		return fmt.Errorf("error found")
	}
	return nil
}

func (h *Holder) FindTx(txHash []byte) (core.Tx, error) {
	return core.Tx{}, nil // TODO complete it
}

func (h *Holder) FindBlockInCache(hash []byte) (*core.Block, error) {
	return nil, nil
}

func (h *Holder) FindTxInCache(txHash []byte) (core.Tx, error) {
	var tx core.Tx
	var txp core.TxPoolImpl
	var err error
	for _, blk := range h.singleBlocks {
		txp.Decode(blk.Content)
		tx, err := txp.Find(txHash)
		if err == nil {
			return tx, err
		}
	}
	found := h.cachedRoot.iterate(func(bct *BlockCacheTree) bool {
		txp.Decode(bct.blk.Content)
		tx, err = txp.Find(txHash)
		if err == nil {
			return true
		} else {
			return false
		}
	})

	if found {
		return tx, err
	} else {
		return tx, fmt.Errorf("not found")
	}
}

func (h *Holder) LongestChain() core.BlockChain {
	rtn := CachedBlockChain{
		BlockChain:  h.bc,
		cachedBlock: make([]*core.Block, 0),
	}
	bct := h.cachedRoot
	if bct.depth == 0 {
		return h.bc
	}
	for {
		if bct.depth == 0 {
			rtn.Push(bct.blk)
			return rtn
		}
		for _, b := range bct.children {
			if b.depth == bct.depth-1 {
				if bct != h.cachedRoot {
					rtn.Push(bct.blk)
				}
				bct = b
				break
			}
		}
	}
}

func (h *Holder) LongestPool() core.UTXOPool {
	bct := h.cachedRoot
	for {
		if bct.depth == 0 {
			return bct.pool
		}
		for _, b := range bct.children {
			if b.depth == bct.depth-1 {
				bct = b
				break
			}
		}
	}
}
