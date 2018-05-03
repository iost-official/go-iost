package core

import "github.com/iost-official/prototype/core"

type TxStatus int

const (
	ACCEPT TxStatus = iota
	CACHED
	POOL
	REJECT
	EXPIRED
)

type Consensus interface {
	Init(bc core.BlockChain, network core.Network) error // ??
	Run()
	Stop()

	GetBlockChain() core.BlockChain
	GetCachedBlockChain() core.BlockChain
}
