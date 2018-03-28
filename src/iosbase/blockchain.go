package iosbase

// Block chain
type BlockChain interface {
	Get(layer int) (Block, error)
	Push(block Block) error
	Length() int

	SubChain(layer int) BlockChain
}
