package core

import (
	"github.com/iost-official/prototype/core/block"
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

	GetBlockChain() block.Chain
	GetCachedBlockChain() block.Chain
}
