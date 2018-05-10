package consensus_common

import (
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/verifier"
)

//go:generate gencode go -schema=structs.schema -package=consensus_common

//// 验证块头正确性，调用此函数时块的父亲节点已经找到
//func VerifyBlockHead(blk *block.Block, parentBlk *block.Block) bool {
//	bh := blk.Head
//	// parent hash
//	if !bytes.Equal(bh.ParentHash, parentBlk.Head.Hash()) {
//		return false
//	}
//	// block number
//	if bh.Number != parentBlk.Head.Number+1 {
//		return false
//	}
//	treeHash := calcTreeHash(DecodeTxs(blk.Content))
//	// merkle tree hash
//	if !bytes.Equal(treeHash, bh.TreeHash) {
//		return false
//	}
//	return true
//}
//
//func calcTreeHash(txs []tx.Tx) []byte {
//	return nil
//}
//
//// 验证块内交易的正确性
//func VerifyBlockContent(blk *block.Block, chain block.Chain) (state.Pool, error) {
//	txs := blk.Content
//	var contracts []vm.Contract
//	for _, tx := range txs {
//		contracts = append(contracts, tx.Contract)
//	}
//	verify := verifier.NewBlockVerifier(chain.GetStatePool())
//	newPool, err := verify.VerifyBlock(*blk, false)
//	if err != nil {
//		return nil, err
//	}
//	return newPool, nil
//}

var ver *verifier.BlockVerifier

func StdBlockVerifier(block *block.Block, pool state.Pool) (state.Pool, error) {
	if ver == nil {
		veri := verifier.NewBlockVerifier(pool)
		ver = &veri
	}
	ver.SetPool(pool)
	return ver.VerifyBlock(block, false)
}

// 验证单个交易的正确性
// 在调用之前需要先调用vm.NewCacheVerifier(pool state.Pool)生成一个cache verifier
// TODO: 考虑自己生成块到达最后一个交易时，直接用返回的state pool更新block cache中的state
func VeirifyTx(tx tx.Tx, cv *verifier.CacheVerifier) (state.Pool, bool) {
	newPool, err := cv.VerifyContract(tx.Contract, false)
	return newPool, err == nil
}

func VerifyTxSig(tx tx.Tx) bool {
	err := tx.VerifySelf()
	return err == nil
}
