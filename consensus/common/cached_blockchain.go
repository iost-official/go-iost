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
	// DPoS中使用，记录该节点被哪些witness确认
	confirmed map[string]int
}

func NewCBC(chain block.Chain) CachedBlockChain {
	return CachedBlockChain{
		Chain:        chain,
		block:        nil,
		parent:       nil,
		cachedLength: 0,
		confirmed:    make(map[string]int),
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
	witness := block.Head.Witness
	c.confirmed[witness] = 1

	confirmed := make(map[string]int)
	confirmed[witness] = 1
	cbc := c.parent
	for cbc.parent != nil {
		witness = cbc.block.Head.Witness
		if _, ok := confirmed[witness]; !ok {
			confirmed[witness] = 0
		}
		confirmed[witness]++
		if len(confirmed) > len(cbc.parent.confirmed) {
			// 如果当前分支有更多的确认，则更新父亲的confirmed
			mapCopy(cbc.parent.confirmed, confirmed)
		} else {
			break
		}
		cbc = cbc.parent
	}
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
		confirmed:    make(map[string]int),
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

func mapCopy(to map[string]int, from map[string]int) {
	for key, value := range from {
		to[key] = value
	}
}
