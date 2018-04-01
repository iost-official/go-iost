package protocol

import (
	"IOS/src/iosbase"
	"fmt"
)

type DatabaseDep interface {
	Init()

	GetIdentity() iosbase.Member
	GetCurrentView() View
	GetBlockChain() iosbase.BlockChain
	GetStatePool() iosbase.StatePool

	PushBlock(block iosbase.Block)

	VerifyTx(tx iosbase.Tx) error
	VerifyBlock(block iosbase.Block) error
}

type RuntimeData struct {
	iosbase.Member

	character  Character
	view       View
	//ExitSignal chan bool

	blockChain iosbase.BlockChain
	statePool  iosbase.StatePool
}

func (d *RuntimeData) VerifyTx(tx iosbase.Tx) error {
	// here only existence of Tx inputs will be verified
	for _, in := range tx.Inputs {
		if _, err := d.statePool.Find(in.StateHash); err == nil {
			return fmt.Errorf("some input not found")
		}
	}
	return nil
}

func (d *RuntimeData) SetView(view View) error {
	d.view = view
	return nil
}