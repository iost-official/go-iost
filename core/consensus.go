package core

type TxStatus int

const (
	ACCEPT TxStatus = iota
	CACHED
	POOL
	REJECT
	EXPIRED
)

type Consensus interface {
	Init(bc BlockChain, network Network) error
	Run()
	Stop()

	GetBlockChain() BlockChain
	GetCachedBlockChain() BlockChain
}
