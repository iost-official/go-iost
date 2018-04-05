package iosbase

type TxStatus int

const (
	ACCEPT TxStatus = iota
	CACHED
	POOL
	REJECT
)

type Consensus interface {
	Init(bc BlockChain, sp StatePool, network Network) error
	Run()
	Stop()

	PublishTx(tx Tx) error
	CheckTx(tx Tx) (TxStatus, error)

	GetStatus() (BlockChain, StatePool, error)
	GetCachedStatus() (BlockChain, StatePool, error)
}
