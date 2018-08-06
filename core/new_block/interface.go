// Package block 是区块和区块链的结构体定义和操作方法
package block

//go:generate mockgen -destination ../mocks/mock_blockchain.go -package core_mock github.com/iost-official/Go-IOS-Protocol/core/block Chain

type Chain interface {
	Push(block *Block) error
	Length() uint64
	CheckLength() error
	Top() *Block // 语法糖
	GetHashByNumber(number uint64) []byte
	GetBlockByNumber(number uint64) *Block
	GetBlockByHash(blockHash []byte) *Block
	GetBlockByteByHash(blockHash []byte) ([]byte, error)
}

type ChainIterator interface {
	Next() *Block
}
