package consensus_common

import (
	"bytes"
	"errors"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/verifier"
)

//go:generate gencode go -schema=structs.schema -package=consensus_common

// VerifyBlockHead 验证块头正确性
// blk 需要验证的块, parentBlk 父块
func VerifyBlockHead(blk *block.Block, parentBlk *block.Block) error {
	bh := blk.Head
	// parent hash
	if !bytes.Equal(bh.ParentHash, parentBlk.Head.Hash()) {
		return errors.New("wrong parent hash")
	}
	// block number
	if bh.Number != parentBlk.Head.Number+1 {
		return errors.New("wrong number")
	}
	//	treeHash := calcTreeHash(DecodeTxs(blk.Content))
	//	// merkle tree hash
	//	if !bytes.Equal(treeHash, bh.TreeHash) {
	//		return false
	//	}
	return nil
}

var ver *verifier.BlockVerifier

// StdBlockVerifier 块内交易的验证函数
func StdBlockVerifier(block *block.Block, pool state.Pool) (state.Pool, error) {
	if ver == nil {
		veri := verifier.NewBlockVerifier(pool)
		ver = &veri
	}
	ver.SetPool(pool)
	return ver.VerifyBlock(block, false)
}

// VeirifyTx 验证单个交易的正确性
// 在调用之前需要先调用vm.NewCacheVerifier(pool state.Pool)生成一个cache verifier
// TODO: 考虑自己生成块到达最后一个交易时，直接用返回的state pool更新block cache中的state
func VeirifyTx(tx tx.Tx, cv *verifier.CacheVerifier) (state.Pool, bool) {
	newPool, err := cv.VerifyContract(tx.Contract, false)
	return newPool, err == nil
}

// VerifyTxSig 验证交易的签名
func VerifyTxSig(tx tx.Tx) bool {
	err := tx.VerifySelf()
	return err == nil
}
