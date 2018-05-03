package block

//go:generate mockgen -destination ../mocks/mock_blockchain.go -package core_mock github.com/iost-official/prototype/core/block BlockChain

// Block chain
type Chain interface {
	Push(block *Block) error // 加入block，检查block是否合法在consensus内实现以解耦合
	Length() int
	Top() *Block // 语法糖

	Iterator() ChainIterator
}

type ChainIterator interface {
	Next() *Block // 返回下一个块
}
