package iosbase

// Block chain
type BlockChain interface {
	Get(layer int) (Block, error) // 获取某层的Block
	Push(block Block) error       // 加入block，检查block是否合法在consensus内实现以解耦合
	Length() int

	SubChain(layer int) BlockChain // 获得从创世区块到某层区块的
}
