package core

import (
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/network"
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
	Init(bc block.Chain, network network.Network) error // ??
	Run()
	Stop()

	GetBlockChain() block.Chain
	GetCachedBlockChain() block.Chain
}
