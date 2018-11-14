package verifier

import (
	"fmt"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
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
	t = tx.NewTx(t.Actions, t.Signers, t.GasLimit, t.GasPrice, t.Expiration, t.Delay)

	txData, err := baseTxData(blk.Head, parent.Head)
	ilog.Info(blk.Head.Number, blk.Head.Witness[4:6], parent.Head, txData)
	if err != nil {
		return nil, err
	}
	if len(t.Actions) > 0 {
		t.Actions[0].Data = txData
	}
	return t, nil
}

func baseTxData(bh *block.BlockHead, pbh *block.BlockHead) (string, error) {
	if pbh != nil {
		ilog.Info(bh.Number, bh.Witness[4:6])
		ilog.Info(pbh)
		return fmt.Sprintf(`[{"parent":["%v", "%v"]}]`, pbh.Witness, pbh.GasUsage), nil
	}
	if bh.Number != 0 {
		return "", fmt.Errorf("block dit not have parent")
	}
	return `[{"parent":["", "0"]}]`, nil
}
