package verifier

import "github.com/iost-official/go-iost/core/tx"

var BlockBaseTx = &tx.Tx{
	Publisher: "_Block_Base",
	Actions:   []*tx.Action{},
}
