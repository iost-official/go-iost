package consensus_common


import (
	"bytes"
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

func calcTreeHash(txs []core.Tx) []byte {
	return nil
}

/*
// 验证块内交易的正确性
func VerifyBlockContent(blk *core.Block, chain core.BlockChain) bool {
	txs := DecodeTxs(blk.Head.BlockHash)
	var contracts []vm.Contract
	for _, tx := range txs {
		contracts = append(contracts, tx.Contract)
	}
	newPool, err := vm.VerifyBlock(contracts, chain.GetStatePool())
	if err != nil {
		return false
	}
	chain.SetStatePool(newPool)
	return true
}

// 验证单个交易的正确性
// 在调用之前需要先调用vm.NewCacheVerifier(pool state.Pool)生成一个cache verifier
// TODO: 考虑自己生成块到达最后一个交易时，直接用返回的state pool更新block cache中的state
func VeirifyTx(tx core.Tx, cv *vm.CacheVerifier) (state.Pool, bool) {
	newPool, err := cv.VerifyContract(tx.Contract, false)
	return newPool, err == nil
}
*/

func VerifyTxSig(tx core.Tx) bool {
	info := make([]byte, 8)
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

func DecodeTxs(content []byte) []core.Tx {
	return nil
}

