package protocol

import "IOS/src/iosbase"

type Database interface {
	NewViewSignal() (chan View, error)
	VerifyTx(tx iosbase.Tx) error
	VerifyBlock(block iosbase.Block) error
	PushBlock(block iosbase.Block) error
	GetStatePool() (iosbase.StatePool, error)
	GetBlockChain() (iosbase.BlockChain, error)
	GetCurrentView() (View, error)
	GetIdentity() (iosbase.Member, error)
}
