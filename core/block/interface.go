package block

import "github.com/iost-official/prototype/core/state"

//go:generate mockgen -destination ../mocks/mock_blockchain.go -package core_mock github.com/iost-official/prototype/core/block Chain

// Block chain
type Chain interface {
	Push(block *Block) error
	Length() uint64
	Top() *Block // 语法糖
	GetBlockByNumber(number uint64) *Block
	GetBlockByHash(blockHash []byte) *Block

	// chain中的state pool相关
	GetStatePool() state.Pool
	SetStatePool(pool state.Pool)

	Iterator() ChainIterator
}

type ChainIterator interface {
	Next() *Block // 返回下一个块
}
