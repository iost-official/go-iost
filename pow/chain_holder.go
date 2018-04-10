package pow

import (
	"github.com/iost-official/prototype/core"
	"bytes"
	"fmt"
)

const (
	MaxCacheDepth = 6
)

type BlockCacheTree struct {
	depth    int
	blk      *core.Block
	children []*BlockCacheTree
	super    *BlockCacheTree
}

type CacheStatus int

const (
	Extend   CacheStatus = iota
	Fork
	NotFound
)

func newBct(block *core.Block, tree *BlockCacheTree) *BlockCacheTree {
	bct := BlockCacheTree{
		depth:    0,
		blk:      block,
		children: make([]*BlockCacheTree, 0),
		super:    tree,
	}
	return &bct
}

func (b *BlockCacheTree) add(block *core.Block) CacheStatus {

	if b.blk == nil {
		b.blk = block
		return Extend
	}

	var code CacheStatus
	for _, bct := range b.children {
		code = bct.add(block)
		if code == Extend {
			if bct.depth == b.depth {
				b.depth ++
				return Extend
			} else {
				return Fork
			}
		} else if code == Fork {
			return Fork
		}
	}

	if bytes.Equal(b.blk.Head.Hash(), block.Head.SuperHash) {
		if len(b.children) == 0 {
			b.children = append(b.children, newBct(block, b))
			b.depth ++
			return Extend
		} else {
			return Fork
		}
	}

	return NotFound
}

func (b *BlockCacheTree) pop() *core.Block {
	blk := b.blk
	for _, bct := range b.children {
		if bct.depth == b.depth-1 {
			b = bct
			b.super = nil
		}
	}
	return blk
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
			blk:      nil,
			children: make([]*BlockCacheTree, 0),
			super:    nil,
		},
		pool:pool,
		singleBlocks:make([]*core.Block, 0),
	}
	return h
}

func (h *Holder) Add(block *core.Block) {
	code := h.cachedRoot.add(block)
	switch code {
	case Extend:
		if h.cachedRoot.depth > MaxCacheDepth {
			h.bc.Push(h.cachedRoot.pop())
		}
		fallthrough
	case Fork:
		for _, blk := range h.singleBlocks {
			h.Add(blk)
		}
	case NotFound:
		h.singleBlocks = append(h.singleBlocks, block)
	}
}

func (h *Holder) FindTx(txHash []byte) (core.Tx, error) {
	return core.Tx{}, nil // TODO complete it
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
