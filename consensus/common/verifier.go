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

var ver *verifier.CacheVerifier

// StdBlockVerifier 块内交易的验证函数
func StdBlockVerifier(block *block.Block, pool state.Pool) (state.Pool, error) {
	txs := block.Content
	ptxs := make([]*tx.Tx, 0)
	for i, _ := range txs {
		ptxs = append(ptxs, &(txs[i]))
	}
	pool2, _, err := StdTxsVerifier(ptxs, pool.Copy())
	if err != nil {
		return pool, err
	}
	return pool2.MergeParent()
}

func StdTxsVerifier(txs []*tx.Tx, pool state.Pool) (state.Pool, int, error) {
	pool2 := pool.Copy()
	for i, txx := range txs {
		var err error
		pool2, err = ver.VerifyContract(txx.Contract, pool2)
		////////////probe//////////////////TODO 严重影响性能，不应该在这里测试
		//var ret string = "pass"
		//if err != nil {
		//	ret = "fail"
		//}
		//log.Report(&log.MsgTx{
		//	SubType:   "verify." + ret,
		//	TxHash:    txx.Hash(),
		//	Publisher: txx.Publisher.Pubkey,
		//	Nonce:     txx.Nonce,
		//})
		///////////////////////////////////

		if err != nil {
			return pool2, i, err
		}

	}

	return pool2, len(txs), nil
}

func CleanStdVerifier() {
	ver.CleanUp()
}

func StdCacheVerifier(txx *tx.Tx, pool state.Pool) error {
	p2, err := ver.VerifyContract(txx.Contract, pool.Copy())
	if err != nil {
		return err
	}
	p2.MergeParent()
	return nil
}

// TODO： 请别用这个了
//func VerifyTx(tx *tx.Tx, txVer *verifier.CacheVerifier) (state.Pool, bool) {
//	newPool, err := txVer.VerifyContract(tx.Contract, false)
//
//	////////////probe//////////////////
//	var ret string = "pass"
//	if err != nil {
//		ret = "fail"
//	}
//	log.Report(&log.MsgTx{
//		SubType:   "verify." + ret,
//		TxHash:    tx.Hash(),
//		Publisher: tx.Publisher.Pubkey,
//		Nonce:     tx.Nonce,
//	})
//	///////////////////////////////////
//
//	return newPool, err == nil
//}

// VerifyTxSig 验证交易的签名
func VerifyTxSig(tx tx.Tx) bool {
	err := tx.VerifySelf()
	return err == nil
}

func init() {
	veri := verifier.NewCacheVerifier()
	ver = &veri
}
