package block

import (
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/core/state"
)

type BlockChainImpl struct {
	db     db.LDBDatabase
	redis  db.RedisDatabase
	length int32
}

func (b *BlockChainImpl) Push(block *Block) error {

	return nil
}

func (b *BlockChainImpl) Length() int {
	return 0
}

func (b *BlockChainImpl) Top() *Block {
	return &Block{}
}

func (b *BlockChainImpl) GetStatePool() state.Pool  {
	return nil
}

func (b *BlockChainImpl) SetStatePool(pool state.Pool)  {

}

func (b *BlockChainImpl) Iterator() ChainIterator {
	return nil
}
