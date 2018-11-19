package verifier

import (
	"fmt"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
)

// NewBaseTx is new baseTx
func NewBaseTx(blk *block.Block, parent *block.Block) (*tx.Tx, error) {
	acts := []*tx.Action{}
	if blk.Head.Number > 0 {
		txData, err := baseTxData(blk.Head, parent.Head)
		if err != nil {
			return nil, err
		}
		act := tx.NewAction("base.iost", "Exec", txData)
		acts = append(acts, act)
	}
	tx := &tx.Tx{
		Publisher: "_Block_Base",
		GasLimit:  1000000,
		GasPrice:  100,
		Actions:   acts,
	}
	return tx, nil
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
