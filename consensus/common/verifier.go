package consensus_common

import (
	"bytes"
	"errors"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
	"github.com/iost-official/Go-IOS-Protocol/common"
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

func VerifyBlockHead(blk *block.Block, parentBlk *block.Block, chainTop *block.Block) error {
	bh := blk.Head
	cur := common.GetCurrentTimestamp().Slot
	if bh.Time > cur+1 {
		return ErrFutureBlk
	}
	if bh.Time < chainTop.Head.Time {
		return ErrOldBlk
	}
	if !bytes.Equal(bh.ParentHash, parentBlk.HeadHash()) {
		return ErrParentHash
	}
	if bh.Number != parentBlk.Head.Number+1 {
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

func VerifyBlockWithVM(blk *block.Block, db *db.MVCCDB) error {
	var receipts []*tx.TxReceipt
	engine := new_vm.NewEngine(&blk.Head, db)
	for _, tx := range blk.Txs {
		receipt, err := engine.Exec(tx)
		if err != nil {
			return err
		}
		receipts = append(receipts, receipt)
	}
	for i, r := range receipts {
		if !bytes.Equal(blk.Receipts[i].Encode(), r.Encode()) {
			return ErrTxReceipt
		}
	}
	return nil
}

var txEngine new_vm.Engine

func VerifyTxBegin(blk *block.Block, db *db.MVCCDB) {
	txEngine = new_vm.NewEngine(&blk.Head, db)
}

func VerifyTx(tx *tx.Tx) (*tx.TxReceipt, error) {
	return txEngine.Exec(tx)
}
