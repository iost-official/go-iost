package iosbase

// Block chain
type BlockChain interface {
	Get(layer int) (Block, error)
	Push(block Block) error
	Length() int
	Top() *Block

	SubChain(layer int) BlockChain
}
