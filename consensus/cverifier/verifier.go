package cverifier

import (
	"bytes"
	"errors"
	"time"

	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm"
)

var (
	errFutureBlk  = errors.New("block from future")
	errOldBlk     = errors.New("block too old")
	errParentHash = errors.New("wrong parent hash")
	errNumber     = errors.New("wrong number")
	errTxHash     = errors.New("wrong txs hash")
	errMerkleHash = errors.New("wrong tx receipt merkle hash")
	errTxReceipt  = errors.New("wrong tx receipt")
	// TxExecTimeLimit the maximum verify execution time of a transaction
	TxExecTimeLimit = 400 * time.Millisecond
)

// VerifyBlockHead verifies the block head.
func VerifyBlockHead(blk *block.Block, parentBlock *block.Block, lib *block.Block) error {
	bh := blk.Head
	if bh.Time > time.Now().UnixNano() {
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
	ilog.Infof("[pob] verifyBlockWithVM start, number: %d, hash = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()))
	engine := vm.NewEngine(blk.Head, db)
	for k, t := range blk.Txs {
		et := TxExecTimeLimit
		if blk.Receipts[k].Status.Code == tx.ErrorTimeout {
			et /= 4
		}
		receipt, err := engine.Exec(t, et)
		if err != nil {
			return err
		}
		if !bytes.Equal(blk.Receipts[k].Encode(), receipt.Encode()) {
			ilog.Errorf("block num: %v , receipt: %v, blk.Receipts[%v]: %v, action name: %v", blk.Head.Number, receipt, k, blk.Receipts[k], blk.Txs[k].Actions[0].ActionName)
			return errTxReceipt
		}
	}
	ilog.Infof("[pob] verifyBlockWithVM end, number: %d, hash = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()))
	return nil
}
