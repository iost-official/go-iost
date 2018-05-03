package consensus_common

import (
	"github.com/iost-official/prototype/core/block"
)

type CachedBlockChain struct {
	block.Chain
	cachedBlock []*block.Block
}

func NewCBC(chain block.Chain) CachedBlockChain {
	return CachedBlockChain{
		Chain:  chain,
		cachedBlock: make([]*block.Block, 0),
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
	c.cachedBlock = append(c.cachedBlock, block)
	return nil
}
func (c *CachedBlockChain) Length() int {
	return c.Chain.Length() + len(c.cachedBlock)
}
func (c *CachedBlockChain) Top() *block.Block {
	l := len(c.cachedBlock)
	if l == 0 {
		return c.Chain.Top()
	}
	return c.cachedBlock[l-1]
}

func (c *CachedBlockChain) Copy() CachedBlockChain {
	cbc := CachedBlockChain{
		Chain:  c.Chain,
		cachedBlock: make([]*block.Block, 0),
	}
	copy(cbc.cachedBlock, c.cachedBlock)
	return cbc
}

func (c *CachedBlockChain) Flush() {
	for _, b := range c.cachedBlock {
		c.Chain.Push(b)
	}
}

func (c *CachedBlockChain) Iterator() block.ChainIterator {
	return nil
}
