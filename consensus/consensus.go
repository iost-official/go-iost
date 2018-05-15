package core

import (
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
)

type TxStatus int

const (
	ACCEPT TxStatus = iota
	CACHED
	POOL
	REJECT
	EXPIRED
)

type Consensus interface {
	Run()
	Stop()

	BlockChain() block.Chain
	CachedBlockChain() block.Chain
	StatePool() state.Pool
	CachedStatePool() state.Pool
}
