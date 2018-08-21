package consensus_common

import (
	"bytes"
	"errors"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
)

var (
	ErrFutureBlk  = errors.New("block from future")
	ErrOldBlk     = errors.New("block too old")
	ErrParentHash = errors.New("wrong parent hash")
	ErrNumber     = errors.New("wrong number")
	ErrTxHash     = errors.New("wrong txs hash")
	ErrMerkleHash = errors.New("wrong tx receipt merkle hash")
	ErrTxReceipt  = errors.New("wrong tx receipt")
)

func VerifyBlockHead(blk *block.Block, parentBlock *block.Block, lib *block.Block) error {
	bh := blk.Head
	if bh.Time > time.Now().Unix()/common.SlotLength+1 {
		return ErrFutureBlk
	}
	if bh.Time <= lib.Head.Time {
		return ErrOldBlk
	}
	if !bytes.Equal(bh.ParentHash, parentBlock.HeadHash()) {
		return ErrParentHash
	}
	if bh.Number != parentBlock.Head.Number+1 {
		return ErrNumber
	}
	if !bytes.Equal(blk.CalculateTxsHash(), bh.TxsHash) {
		return ErrTxHash
	}
	if !bytes.Equal(blk.CalculateMerkleHash(), bh.MerkleHash) {
		return ErrMerkleHash
	}
	return nil
}

func VerifyBlockWithVM(blk *block.Block, db db.MVCCDB) error {
	engine := new_vm.NewEngine(&blk.Head, db)
	for k, tx := range blk.Txs {
		receipt, err := engine.Exec(tx)
		if err != nil {
			return err
		}
		if !bytes.Equal(blk.Receipts[k].Encode(), receipt.Encode()) {
			return ErrTxReceipt
		}
	}
	return nil
}
