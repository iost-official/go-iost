package consensus_common

import (
	"bytes"
	"errors"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
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

func VerifyBlock(blk *block.Block, commit string) (string, error) {
	var receipts []tx.TxReceipt
	for i := range blk.Txs {
		newCommit, receipt, err := VerifyTx(&blk.Txs[i], commit, &blk.Head)
		if err == nil {
			commit = newCommit
			receipts = append(receipts, receipt)
		} else {
			return "", err
		}
	}
	for i := range receipts {
		if !blk.Receipts[i].Equal(receipts[i]) {
			return "", errors.New("wrong tx receipt")
		}
	}
	return commit, nil
}

func VerifyTx(tx *tx.Tx, commit string, head *block.BlockHead) (string, tx.TxReceipt, error) {
	engine := new_vm.Engine()
	engine.SetEnv(head, commit)
	receipt, newCommit, err := engine.Exec(*tx)
	return newCommit, receipt, err
}
