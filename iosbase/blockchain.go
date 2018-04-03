package iosbase

//go:generate mockgen -destination mocks/mock_blockchain.go -package iosbase_mock -source blockchain.go -imports .=github.com/iost-official/PrototypeWorks/iosbase
// Block chain
type BlockChain interface {
	Get(layer int) (Block, error)
	Push(block Block) error
	Length() int
	Top() *Block

	SubChain(layer int) BlockChain
}
