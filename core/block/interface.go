package block

//go:generate mockgen -destination ../mocks/mock_blockchain.go -package core_mock github.com/iost-official/prototype/core/block Chain

// Block chain
type Chain interface {
	Push(block *Block) error
	Length() uint64
	Top() *Block // 语法糖
	GetBlockByNumber(number uint64) *Block
	GetBlockByHash(blockHash []byte) *Block

	Iterator() ChainIterator
}

type ChainIterator interface {
	Next() *Block // 返回下一个块
}
