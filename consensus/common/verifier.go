package consensus_common

/*
import (
	"bytes"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/core"
	"encoding/binary"
	"github.com/iost-official/prototype/common"
)

// 验证块头正确性，调用此函数时块的父亲节点已经找到
func VerifyBlockHead(blk *core.Block, parentBlk *core.Block) bool {
	bh := blk.Head
	// parent hash
	if !bytes.Equal(bh.ParentHash, parentBlk.Head.Hash()) {
		return false
	}
	// block number
	if bh.Number != parentBlk.Head.Number + 1 {
		return false
	}
	treeHash := calcTreeHash(DecodeTxs(blk.Content))
	// merkle tree hash
	if !bytes.Equal(treeHash, bh.TreeHash) {
		return false
	}
	return true
}

func calcTreeHash(txs []Tx) []byte {
	return nil
}

// 验证块内交易的正确性
func VerifyBlockContent(blk *core.Block, chain core.BlockChain) bool {
	txs := DecodeTxs(blk.content)
	return vm.ExecBlockTxs(txs, chain)
}

// 验证单个交易的正确性
func VeirifyTx(tx core.Tx, chain core.BlockChain) bool {
	return !vm.ExecCachedTxs(tx, chain)
}

func VerifyTxSig(tx core.Tx) bool {
	var info []byte
	binary.BigEndian.PutUint64(info, uint64(tx.Time))
	info = append(info, tx.Contract.Encode()...)
	for _, sign := range tx.Signs {
		if !common.VerifySignature(info, sign) {
			return false
		}
	}
	for _, sign := range tx.Signs {
		info = append(info, sign.Encode()...)
	}
	for _, sign := range tx.Publisher {
		if !common.VerifySignature(info, sign) {
			return false
		}
	}
	return true
}
*/
