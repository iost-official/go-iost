package blockcache

import (
	"github.com/iost-official/prototype/core/block"

	"bytes"
	"github.com/iost-official/prototype/log"
)

type CachedBlockChain struct {
	block.Chain
	block        *block.Block
	cachedLength int
	parent       *CachedBlockChain
	depth        int
	confirmed    int
}

func NewCBC(chain block.Chain) CachedBlockChain {
	return CachedBlockChain{
		Chain:        chain,
		block:        nil,
		parent:       nil,
		cachedLength: 0,
		depth:        0,
		confirmed:    0,
	}
}

func (c *CachedBlockChain) Push(block *block.Block) error {
	c.block = block
	c.cachedLength++

	switch block.Head.Version {
	case 0:
		c.confirmed = 1
		witness := block.Head.Witness
		confirmed := make(map[string]int)
		confirmed[witness] = 1
		cbc := c
		for cbc.parent != nil {
			witness = cbc.parent.Top().Head.Witness
			if _, ok := confirmed[witness]; !ok {
				confirmed[witness] = 0
			}
			confirmed[witness]++
			if len(confirmed) > cbc.parent.confirmed {
				cbc.parent.confirmed = len(confirmed)
			}
			cbc = cbc.parent
		}
		fallthrough
	case 1:
		c.depth = 0
		cbc := c
		depth := 0
		for cbc.parent != nil {
			depth++
			if depth > cbc.parent.depth {
				cbc.parent.depth = depth
			}
			cbc = cbc.parent
		}
	}

	return nil
}

func (c *CachedBlockChain) Length() uint64 {
	return c.Chain.Length() + uint64(c.cachedLength)
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
		confirmed:    0,
	}
	return cbc
}

func (c *CachedBlockChain) Flush() {
	if c.block != nil {
		err := c.Chain.Push(c.block)
		if err != nil {
			log.Log.E("Database error, CachedBlockChain Flush err:%v", err)
		}
		c.block = nil
		c.cachedLength = 0
		c.parent = nil
	}
}

func (c *CachedBlockChain) Iterator() block.ChainIterator {
	return &CBCIterator{c, nil}
}

type CBCIterator struct {
	pc       *CachedBlockChain
	iterator block.ChainIterator
}

func (ci *CBCIterator) Next() *block.Block {
	if ci.iterator != nil {
		return ci.iterator.Next()
	}
	p := ci.pc.block
	if ci.pc.parent == nil {
		ci.iterator = ci.pc.Chain.Iterator()
	}
	ci.pc = ci.pc.parent
	return p
}

func (c *CachedBlockChain) GetBlockByNumber(number uint64) *block.Block {
	if number < c.Chain.Length() {
		return c.Chain.GetBlockByNumber(number)
	}
	if number >= c.Length() {
		return nil
	}
	cbc := c
	for cbc.block != nil {
		if uint64(cbc.block.Head.Number) == number {
			return cbc.block
		}
		cbc = cbc.parent
	}
	return nil
}

func (c *CachedBlockChain) GetHashByNumber(number uint64) []byte {
	if number < c.Chain.Length() {
		return c.Chain.GetHashByNumber(number)
	}
	if number >= c.Length() {
		return nil
	}
	cbc := c
	for cbc.block != nil {
		if uint64(cbc.block.Head.Number) == number {
			return cbc.block.HeadHash()
		}
		cbc = cbc.parent
	}
	return nil
}

func (c *CachedBlockChain) GetBlockByHash(blockHash []byte) *block.Block {
	cbc := c
	for cbc != nil && cbc.block != nil {
		if bytes.Equal(cbc.block.HeadHash(), blockHash) {
			return cbc.block
		}
		cbc = cbc.parent
	}
	return c.Chain.GetBlockByHash(blockHash)
}
