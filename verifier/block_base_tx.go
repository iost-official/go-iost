package verifier

import (
	"fmt"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
)

// BlockBaseTx the first tx in a block
var BlockBaseTx = &tx.Tx{
	Publisher: "_Block_Base",
	GasLimit:  1000000,
	GasPrice:  100,
	Actions:   []*tx.Action{},
}

// NewBaseTx is new baseTx
func NewBaseTx(blk *block.Block, parent *block.Block) (*tx.Tx, error) {
	t := BlockBaseTx
	if len(t.Actions) == 0 {
		return t, nil
	}
	txData, err := baseTxData(blk.Head, parent.Head)
	if err != nil {
		return nil, err
	}
	// manually deep copy actions
	act := *BlockBaseTx.Actions[0]
	act.Data = txData
	acts := []*tx.Action{&act}
	t = tx.NewTx(acts, t.Signers, t.GasLimit, t.GasPrice, t.Expiration, t.Delay)
	return t, nil
}

func baseTxData(bh *block.BlockHead, pbh *block.BlockHead) (string, error) {
	if pbh != nil {
		return fmt.Sprintf(`[{"parent":["%v", "%v"]}]`, pbh.Witness, pbh.GasUsage), nil
	}
	if bh.Number != 0 {
		return "", fmt.Errorf("block dit not have parent")
	}
	return `[{"parent":["", "0"]}]`, nil
}
