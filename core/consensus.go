package core

type TxStatus int

const (
	ACCEPT TxStatus = iota
	CACHED
	POOL
	REJECT
)

type Consensus interface {
	Init(bc BlockChain, sp UTXOPool, network Network) error
	Run()
	Stop()

	PublishTx(tx Tx) error
	CheckTx(tx Tx) (TxStatus, error)

	GetStatus() (BlockChain, UTXOPool, error)
	GetCachedStatus() (BlockChain, UTXOPool, error)
}
