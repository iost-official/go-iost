package blockcache

import (
	"bytes"
	"errors"

	"sync"

	"time"

	"github.com/flybikeGx/easy-timeout/timelimit"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/verifier"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/host"
)

//go:generate gencode go -schema=structs.schema -package=block

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

var blockLock sync.Mutex

func StdBlockVerifier(block *block.Block, pool state.Pool) (state.Pool, error) {
	blockLock.Lock()
	defer blockLock.Unlock()
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

		if err != nil {
			return pool2, i, err
		}

	}

	return pool2, len(txs), nil
}

func CleanStdVerifier() {
	verb.CleanUp()
}

var cacheLock sync.Mutex

func StdCacheVerifier(txx *tx.Tx, pool state.Pool, context *vm.Context) error {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	verb.Context = context

	var p2 state.Pool = nil
	var err error = nil

	if timelimit.Run(200*time.Millisecond, func() {
		defer func() {
			if err0 := recover(); err0 != nil {
				err = err0.(error)
			}
		}()
		p2, err = verb.VerifyContract(txx.Contract, pool.Copy())
	}) {
		if err != nil {
			host.Log(err.Error(), txx.Contract.Info().Prefix)
			return err
		}
		p2.MergeParent()
		return nil
	} else {
		return errors.New("time out")
	}
}

type VerifyContext struct {
	VParentHash []byte
}

func (v VerifyContext) ParentHash() []byte {
	return v.VParentHash
}

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
