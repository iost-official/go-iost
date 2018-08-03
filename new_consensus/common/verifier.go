package consensus_common

import (
	"bytes"
	"errors"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
	"github.com/iost-official/Go-IOS-Protocol/db"
)

func VerifyBlockHead(blk *block.Block, parentBlk *block.Block) error {
	bh := blk.Head
	// time
	cur := GetCurrentTimestamp().Slot
	if bh.Time > cur+1 {
		return errors.New("block from future")
	}
	if bh.Time < block.Chain.Top().Head.Time {
		return errors.New("block too old")
	}
	// parent hash
	if !bytes.Equal(bh.ParentHash, parentBlk.HeadHash()) {
		return errors.New("wrong parent hash")
	}
	// block number
	if bh.Number != parentBlk.Head.Number+1 {
		return errors.New("wrong number")
	}
	// tx hash
	txHash := blk.CalculateTxsHash()
	if !bytes.Equal(txHash, bh.TxsHash) {
		return errors.New("wrong txs hash")
	}
	// tx receipt merkle hash
	merkleHash := blk.CalculateMerkleHash()
	if !bytes.Equal(merkleHash, bh.MerkleHash) {
		return errors.New("wrong tx receipt merkle hash")
	}
	return nil
}

func VerifyBlock(blk *block.Block, db *db.MVCCDB) error {
	var receipts []tx.TxReceipt
	engine := new_vm.NewEngine(blk.Head, db)
	for i := range blk.Txs {
		receipt, err := verify(&blk.Txs[i], &engine)
		if err == nil {
			// commit on every tx? with what tag? what about the last?
			db.Commit()
			receipts = append(receipts, receipt)
		} else {
			db.Rollback()
			return err
		}
	}
	for i := range receipts {
		if !blk.Receipts[i].Equal(receipts[i]) {
			// How to rollback all the txs?
			db.Rollback()
			return errors.New("wrong tx receipt")
		}
	}
	return nil
}

var txEngine new_vm.Engine

func VerifyTxBegin(blk *block.Block, db *db.MVCCDB) {
	txEngine = new_vm.NewEngine(blk.Head, db)
}

func VerifyTx(tx *tx.Tx) (tx.TxReceipt, error) {
	return verify(tx, &txEngine)
}

func verify(tx *tx.Tx, engine *new_vm.Engine) (tx.TxReceipt, error) {
	receipt, err := engine.Exec(tx)
	return receipt, err
}
