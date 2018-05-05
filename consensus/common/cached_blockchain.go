package consensus_common

import (
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
)

type CachedBlockChain struct {
	block.Chain
	block        *block.Block
	pool         state.Pool
	cachedLength int
	parent       *CachedBlockChain
}

func NewCBC(chain block.Chain) CachedBlockChain {
	return CachedBlockChain{
		Chain:        chain,
		block:        nil,
		parent:       nil,
		cachedLength: 0,
	}
}

//func (c *CachedBlockChain) Get(layer int) (*core.Block, error) {
//	if layer < 0 || layer >= c.BlockChain.Length()+len(c.cachedBlock) {
//		return nil, fmt.Errorf("overflow")
//	}
//	if layer < c.BlockChain.Length() {
//		return c.BlockChain.Get(layer)
//	}
//	return c.cachedBlock[layer-c.BlockChain.Length()], nil
//}
func (c *CachedBlockChain) Push(block *block.Block) error {
	c.block = block
	c.cachedLength++
	return nil
}
func (c *CachedBlockChain) Length() int {
	return c.Chain.Length() + c.cachedLength
}
func (c *CachedBlockChain) Top() *block.Block {
	if c.cachedLength == 0 {
		return c.Chain.Top()
	}
	return c.block
}

func (c *CachedBlockChain) Copy() CachedBlockChain {
	cbc := CachedBlockChain{
		Chain:        c.Chain,
		parent:       c,
		cachedLength: c.cachedLength,
		pool:         c.pool,
	}
	return cbc
}

// 调用时保证只flush未确认块的第一个，如果要flush多个，需多次调用Flush()
func (c *CachedBlockChain) Flush() {
	if c.block != nil {
		c.Chain.Push(c.block)
		//TODO: chain实现后去掉注释
		//c.Chain.SetStatePool(c.pool)
		c.block = nil
		c.cachedLength = 0
		c.parent = nil
	}
}

func (c *CachedBlockChain) GetStatePool() state.Pool {
	return c.pool
}

func (c *CachedBlockChain) SetStatePool(pool state.Pool) {
	c.pool = pool
}

func (c *CachedBlockChain) Iterator() block.ChainIterator {
	return nil
}
