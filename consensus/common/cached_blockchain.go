package consensus_common

import (
	"github.com/iost-official/prototype/core"
)

type CachedBlockChain struct {
	core.BlockChain
	cachedBlock []*core.Block
}

func NewCBC(chain core.BlockChain) CachedBlockChain {
	return CachedBlockChain{
		BlockChain:  chain,
		cachedBlock: make([]*core.Block, 0),
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

func (c *CachedBlockChain) Push(block *core.Block) error {
	c.cachedBlock = append(c.cachedBlock, block)
	return nil
}
func (c *CachedBlockChain) Length() int {
	return c.BlockChain.Length() + len(c.cachedBlock)
}
func (c *CachedBlockChain) Top() *core.Block {
	l := len(c.cachedBlock)
	if l == 0 {
		return c.BlockChain.Top()
	}
	return c.cachedBlock[l-1]
}

func (c *CachedBlockChain) Copy() CachedBlockChain {
	cbc := CachedBlockChain{
		BlockChain:  c.BlockChain,
		cachedBlock: make([]*core.Block, 0),
	}
	copy(cbc.cachedBlock, c.cachedBlock)
	return cbc
}

func (c *CachedBlockChain) Flush() {
	for _, b := range c.cachedBlock {
		c.BlockChain.Push(b)
	}
}

func (c *CachedBlockChain) Iterator() core.BlockChainIterator {
	return nil
}
