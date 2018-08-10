package consensus_common

import (
	"bytes"
	"errors"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
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

func VerifyBlockHead(blk *block.Block, parentBlk *block.Block, chainTop *block.Block) error {
	bh := blk.Head
	// time
	cur := GetCurrentTimestamp().Slot
	if bh.Time > cur+1 {
		return ErrFutureBlk
	}
	if bh.Time < chainTop.Head.Time {
		return ErrOldBlk
	}
	// parent hash
	if !bytes.Equal(bh.ParentHash, parentBlk.HeadHash()) {
		return ErrParentHash
	}
	// block number
	if bh.Number != parentBlk.Head.Number+1 {
		return ErrNumber
	}
	// tx hash
	txHash := blk.CalculateTxsHash()
	if !bytes.Equal(txHash, bh.TxsHash) {
		return ErrTxHash
	}
	// tx receipt merkle hash
	merkleHash := blk.CalculateMerkleHash()
	if !bytes.Equal(merkleHash, bh.MerkleHash) {
		return ErrMerkleHash
	}
	return nil
}

func VerifyBlock(blk *block.Block, db *db.MVCCDB) error {
	var receipts []*tx.TxReceipt
	engine := new_vm.NewEngine(&blk.Head, db)
	for _, tx := range blk.Txs {
		receipt, err := verify(tx, engine)
		if err == nil {
			db.Commit()
			receipts = append(receipts, receipt)
		} else {
			db.Rollback()
			return err
		}
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
	return verify(tx, txEngine)
}

func verify(tx *tx.Tx, engine new_vm.Engine) (*tx.TxReceipt, error) {
	receipt, err := engine.Exec(tx)
	return receipt, err
}
