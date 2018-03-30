package protocol

import (
	"IOS/src/iosbase"
	"fmt"
)

type RuntimeData struct {
	iosbase.Member

	character Character
	view      View
	phase     Phase
	isRunning bool

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
