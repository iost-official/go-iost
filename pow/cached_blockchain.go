package pow

import (
	"github.com/iost-official/prototype/core"
	"fmt"
)

type CachedBlockChain struct {
	core.BlockChain
	cachedBlock []*core.Block
}

func (c *CachedBlockChain)Get(layer int) (*core.Block, error){
	if layer < 0 || layer >= c.BlockChain.Length() + len(c.cachedBlock) {
		return nil, fmt.Errorf("overflow")
	}
	if layer < c.BlockChain.Length() {
		return c.BlockChain.Get(layer)
	}
	return c.cachedBlock[layer - c.BlockChain.Length()], nil
}
func (c *CachedBlockChain)Push(block *core.Block) error{
	c.cachedBlock = append(c.cachedBlock, block)
}
func (c *CachedBlockChain)Length() int{
	return c.BlockChain.Length() + len(c.cachedBlock)
}
func (c *CachedBlockChain)Top() *core.Block{
	return c.cachedBlock[len(c.cachedBlock) - 1]
}

func (c *CachedBlockChain)Flush() {
	for _, b := range c.cachedBlock {
		c.BlockChain.Push(b)
	}
}


