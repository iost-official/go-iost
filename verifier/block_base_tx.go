package verifier

import "github.com/iost-official/go-iost/core/tx"

// BlockBaseTx the first tx in a block
var BlockBaseTx = &tx.Tx{
	Publisher: "_Block_Base",
	Actions:   []*tx.Action{},
}
