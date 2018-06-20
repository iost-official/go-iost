package blockcache

import (
	"bytes"
	"errors"

	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/verifier"
	"github.com/iost-official/prototype/vm"
)

//go:generate gencode go -schema=structs.schema -package=block

// VerifyBlockHead 验证块头正确性
// blk 需要验证的块, parentBlk 父块
func VerifyBlockHead(blk *block.Block, parentBlk *block.Block) error {
	bh := blk.Head
	// parent hash
	if !bytes.Equal(bh.ParentHash, parentBlk.HeadHash()) {
		return errors.New("wrong parent hash")
	}
	// block number
	if bh.Number != parentBlk.Head.Number+1 {
		return errors.New("wrong number")
	}
	treeHash := blk.CalculateTreeHash()
	if !bytes.Equal(treeHash, bh.TreeHash) {
		return errors.New("wrong tree hash")
	}
	return nil
}

var ver *verifier.CacheVerifier
var verb *verifier.CacheVerifier

// StdBlockVerifier 块内交易的验证函数
func StdBlockVerifier(block *block.Block, pool state.Pool) (state.Pool, error) {

	ver.Context = vm.NewContext(vm.BaseContext())
	ver.Context.ParentHash = block.Head.ParentHash
	ver.Context.Timestamp = block.Head.Time
	ver.Context.BlockHeight = block.Head.Number
	ver.Context.Witness = vm.IOSTAccount(block.Head.Witness)

	txs := block.Content
	ptxs := make([]*tx.Tx, 0)
	for i := range txs {
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
	verb.CleanUp()
}

func StdCacheVerifier(txx *tx.Tx, pool state.Pool, context *vm.Context) error {
	verb.Context = context

	p2, err := verb.VerifyContract(txx.Contract, pool.Copy())
	if err != nil {
		return err
	}
	p2.MergeParent()
	return nil
}

type VerifyContext struct {
	VParentHash []byte
}

func (v VerifyContext) ParentHash() []byte {
	return v.VParentHash
}

// VerifyTxSig 验证交易的签名
func VerifyTxSig(tx tx.Tx) bool {
	err := tx.VerifySelf()
	return err == nil
}

func init() {
	veri := verifier.NewCacheVerifier()
	ver = &veri

	veri = verifier.NewCacheVerifier()
	verb = &veri
}
