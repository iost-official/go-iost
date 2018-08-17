// Package block 是区块和区块链的结构体定义和操作方法
package block

//go:generate mockgen -destination ../mocks/mock_blockchain.go -package core_mock github.com/iost-official/Go-IOS-Protocol/core/new_block Chain

type Chain interface {
	Push(block *Block) error
	Length() int64
	CheckLength() error
	Top() (*Block, error)
	GetHashByNumber(number int64) ([]byte, error)
	GetBlockByNumber(number int64) (*Block, error)
	GetBlockByHash(blockHash []byte) (*Block, error)
	GetBlockByteByHash(blockHash []byte) ([]byte, error)
}

type ChainIterator interface {
	Next() *Block
}
