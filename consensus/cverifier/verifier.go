package cverifier

import (
	"bytes"
	"errors"
	"time"

	"github.com/iost-official/go-iost/core/block"
)

var (
	errFutureBlk   = errors.New("block from future")
	errParentHash  = errors.New("wrong parent hash")
	errNumber      = errors.New("wrong number")
	errTxHash      = errors.New("wrong txs hash")
	errMerkleHash  = errors.New("wrong tx receipt merkle hash")
	errInvalidTime = errors.New("block time less than parent block")
	// errTxReceipt  = errors.New("wrong tx receipt")

	// TxExecTimeLimit the maximum verify execution time of a transaction
	TxExecTimeLimit = 400 * time.Millisecond

	// MaxBlockTimeGap is the limit of the difference of block time and local time.
	MaxBlockTimeGap = 1 * time.Second.Nanoseconds()
)

// VerifyBlockHead verifies the block head.
func VerifyBlockHead(blk *block.Block, parentBlock *block.Block, lib *block.Block) error {
	bh := blk.Head
	if bh.Time > time.Now().UnixNano()+MaxBlockTimeGap {
		return errFutureBlk
	}
	if bh.Time <= parentBlock.Head.Time {
		return errInvalidTime
	}
	if !bytes.Equal(bh.ParentHash, parentBlock.HeadHash()) {
		return errParentHash
	}
	if bh.Number != parentBlock.Head.Number+1 {
		return errNumber
	}
	if !bytes.Equal(blk.CalculateTxMerkleHash(), bh.TxMerkleHash) {
		return errTxHash
	}
	if !bytes.Equal(blk.CalculateTxReceiptMerkleHash(), bh.TxReceiptMerkleHash) {
		return errMerkleHash
	}

	return nil
}
