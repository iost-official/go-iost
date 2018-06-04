package consensus_common

import (
	"github.com/iost-official/prototype/core/block"

	"bytes"
)

// CachedBlockChain 代表缓存的链
type CachedBlockChain struct {
	block.Chain
	block        *block.Block
	cachedLength int
	parent       *CachedBlockChain
	depth        int
	// PoB中使用，记录该节点被最多几个witness确认
	confirmed int
}

// NewCBC 新建一个缓存链
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

//func (c *CachedBlockChain) Get(layer int) (*core.Block, error) {
//	if layer < 0 || layer >= c.BlockChain.Length()+len(c.cachedBlock) {
//		return nil, fmt.Errorf("overflow")
//	}
//	if layer < c.BlockChain.Length() {
//		return c.BlockChain.Get(layer)
//	}
//	return c.cachedBlock[layer-c.BlockChain.Length()], nil
//}

// Push 把新块加入缓存链尾部
func (c *CachedBlockChain) Push(block *block.Block) error {
	c.block = block
	c.cachedLength++

	// push的时候更新共识相关变量
	switch block.Head.Version {
	case 0:
		// PoB
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
				// 如果当前分支有更多的确认，则更新父亲的confirmed
				cbc.parent.confirmed = len(confirmed)
			}
			cbc = cbc.parent
		}
		fallthrough
	case 1:
		// PoW
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

// Length 返回缓存链的总长度
func (c *CachedBlockChain) Length() uint64 {
	return c.Chain.Length() + uint64(c.cachedLength)
}

// Top 返回缓存链的最新块
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

// Flush 把缓存的链存入数据库
// 调用时保证只flush未确认块的第一个，如果要flush多个，需多次调用Flush()
func (c *CachedBlockChain) Flush() {
	if c.block != nil {
		//[HowHsu]:I think this operation should be done in a new goroutine
		c.Chain.Push(c.block)
		c.block = nil
		c.cachedLength = 0
		c.parent = nil
	}
}

// Iterator 生成一个链迭代器
func (c *CachedBlockChain) Iterator() block.ChainIterator {
	return &CBCIterator{c, nil}
}

// CBCIterator 缓存链的迭代器
type CBCIterator struct {
	pc       *CachedBlockChain
	iterator block.ChainIterator
}

// Next 返回下一个块
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

// GetBlockByNumber 从缓存链里找对应块号的块
// deprecate : 请使用iterator
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

// GetBlockByHash 从缓存链里找对应hash的块
// deprecate : 请使用iterator
func (c *CachedBlockChain) GetBlockByHash(blockHash []byte) *block.Block {
	cbc := c
	for cbc.block != nil {
		if bytes.Equal(cbc.block.Head.Hash(), blockHash) {
			return cbc.block
		}
		cbc = cbc.parent
	}
	return c.Chain.GetBlockByHash(blockHash)
}
