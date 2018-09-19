package verifier

import (
	"bytes"
	"errors"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm"
)

var (
	errFutureBlk  = errors.New("block from future")
	errOldBlk     = errors.New("block too old")
	errParentHash = errors.New("wrong parent hash")
	errNumber     = errors.New("wrong number")
	errTxHash     = errors.New("wrong txs hash")
	errMerkleHash = errors.New("wrong tx receipt merkle hash")
	errTxReceipt  = errors.New("wrong tx receipt")
)

// VerifyBlockHead verifies the block head.
func VerifyBlockHead(blk *block.Block, parentBlock *block.Block, lib *block.Block) error {
	bh := blk.Head
	if bh.Time > time.Now().Unix()/common.SlotLength+1 {
		return errFutureBlk
	}
	if bh.Time < lib.Head.Time {
		return errOldBlk
	}
	if !bytes.Equal(bh.ParentHash, parentBlock.HeadHash()) {
		return errParentHash
	}
	if bh.Number != parentBlock.Head.Number+1 {
		return errNumber
	}
	if !bytes.Equal(blk.CalculateTxsHash(), bh.TxsHash) {
		return errTxHash
	}
	if !bytes.Equal(blk.CalculateMerkleHash(), bh.MerkleHash) {
		return errMerkleHash
	}
	return nil
}

//VerifyBlockWithVM verifies the block with VM.
func VerifyBlockWithVM(blk *block.Block, db db.MVCCDB) error {
	engine := vm.NewEngine(blk.Head, db)
	for k, tx := range blk.Txs {
		receipt, err := engine.Exec(tx)
		if err != nil {
			return err
		}
		if !bytes.Equal(blk.Receipts[k].Encode(), receipt.Encode()) {
			ilog.Errorf("block num: %v , receipt: %v, blk.Receipts[%v]: %v, action name: %v", blk.Head.Number, receipt, k, blk.Receipts[k], blk.Txs[k].Actions[0].ActionName)
			return errTxReceipt
		}
	}
	return nil
}
