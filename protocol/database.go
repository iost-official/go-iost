package protocol

import (
	"github.com/iost-official/PrototypeWorks/iosbase"
)

type Database interface {
	NewViewSignal() (chan View, error)

	VerifyTx(tx iosbase.Tx) error
	VerifyTxWithCache(tx iosbase.Tx, cachePool iosbase.TxPool) error
	VerifyBlock(block iosbase.Block) error
	VerifyBlockWithCache(block *iosbase.Block, cachePool iosbase.TxPool) error

	PushBlock(block *iosbase.Block) error

	GetStatePool() (iosbase.StatePool, error)
	GetBlockChain() (iosbase.BlockChain, error)
	GetCurrentView() (View, error)
	GetIdentity() (iosbase.Member, error)
}
