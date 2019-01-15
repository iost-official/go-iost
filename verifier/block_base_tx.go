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
		txData, err := baseTxData(blk, parent)
		if err != nil {
			return nil, err
		}
		act := tx.NewAction("base.iost", "exec", txData)
		acts = append(acts, act)
	}
	tx := &tx.Tx{
		Publisher: "base.iost",
		GasLimit:  100000000,
		GasRatio:  100,
		Actions:   acts,
		Time:      blk.Head.Time,
		ChainID:   tx.ChainID,
	}
	return tx, nil
}

func baseTxData(b *block.Block, pb *block.Block) (string, error) {
	if pb != nil {
		return fmt.Sprintf(`[{"parent":["%v", "%v"]}]`, pb.Head.Witness, pb.CalculateGasUsage()), nil
	}
	if b.Head.Number != 0 {
		return "", fmt.Errorf("block dit not have parent")
	}
	return `[{"parent":["", "0"]}]`, nil
}
