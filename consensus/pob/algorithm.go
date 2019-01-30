package pob

import (
	"errors"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/cverifier"
	"github.com/iost-official/go-iost/consensus/snapshot"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/verifier"
)

var (
	errWitness                = errors.New("wrong witness")
	errSignature              = errors.New("wrong signature")
	errTxDup                  = errors.New("duplicate tx")
	errDoubleTx               = errors.New("double tx in block")
	errTxLenUnmatchReceiptLen = errors.New("tx len unmatch receipt len")
	generateTxsNum            = 0
)

func generateBlock(
	acc *account.KeyPair,
	txPool txpool.TxPool,
	db db.MVCCDB,
	limitTime time.Duration,
	pTx *txpool.SortedTxMap,
	head *blockcache.BlockCacheNode) (*block.Block, error) {

	ilog.Debug("generate Block start")
	st := time.Now()
	topBlock := head.Block
	blk := &block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: topBlock.HeadHash(),
			Info:       make([]byte, 0),
			Number:     topBlock.Head.Number + 1,
			Witness:    acc.ReadablePubkey(),
			Time:       time.Now().UnixNano(),
		},
		Txs:      []*tx.Tx{},
		Receipts: []*tx.TxReceipt{},
	}
	db.Checkout(string(topBlock.HeadHash()))

	// call vote
	v := verifier.Verifier{}
	t1 := time.Now()
	// TODO: stateDb and block head is consisdent, pTx may be inconsisdent.
	dropList, _, err := v.Gen(blk, topBlock, &head.WitnessList, db, pTx, &verifier.Config{
		Mode:        0,
		Timeout:     limitTime - time.Now().Sub(st),
		TxTimeLimit: common.MaxTxTimeLimit,
	})
	t2 := time.Since(t1)
	if len(blk.Txs) != 0 {
		ilog.Debugf("time spent per tx: %v", t2.Nanoseconds()/int64(len(blk.Txs)))
	}
	if err != nil {
		go txPool.DelTxList(dropList)
		ilog.Errorf("Gen is err: %v", err)
		return nil, err
	}
	blk.Head.TxMerkleHash = blk.CalculateTxMerkleHash()
	blk.Head.TxReceiptMerkleHash = blk.CalculateTxReceiptMerkleHash()
	err = blk.CalculateHeadHash()
	if err != nil {
		return nil, err
	}
	blk.Sign = acc.Sign(blk.HeadHash())
	err = snapshot.Save(db, blk)
	if err != nil {
		return nil, err
	}
	db.Commit(string(blk.HeadHash()))
	metricsGeneratedBlockCount.Add(1, nil)
	generateTxsNum += len(blk.Txs)
	return blk, nil
}

func verifyBasics(blk *block.Block, signature *crypto.Signature) error {
	signature.SetPubkey(account.DecodePubkey(blk.Head.Witness))
	hash := blk.HeadHash()
	if !signature.Verify(hash) {
		return errSignature
	}
	if len(blk.Txs) != len(blk.Receipts) {
		return errTxLenUnmatchReceiptLen
	}
	return nil
}

func verifyBlock(blk, parent, lib *block.Block, witnessList *blockcache.WitnessList, txPool txpool.TxPool, db db.MVCCDB, chain block.Chain, replay bool) error {
	err := cverifier.VerifyBlockHead(blk, parent, lib)
	if err != nil {
		return err
	}

	if replay == false && witnessOfNanoSec(blk.Head.Time, witnessList.Active()) != blk.Head.Witness {
		ilog.Errorf("blk num: %v, time: %v, witness: %v, witness len: %v, witness list: %v",
			blk.Head.Number, blk.Head.Time, blk.Head.Witness, len(witnessList.Active()), witnessList.Active())
		return errWitness
	}
	ilog.Debugf("[pob] start to verify block if foundchain, number: %v, hash = %v, witness = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()), blk.Head.Witness[4:6])
	blkTxSet := make(map[string]bool, len(blk.Txs))
	for i, t := range blk.Txs {
		if blkTxSet[string(t.Hash())] {
			return errDoubleTx
		}
		blkTxSet[string(t.Hash())] = true

		if i == 0 {
			// base tx
			continue
		}
		exist := txPool.ExistTxs(t.Hash(), parent)
		switch exist {
		case txpool.FoundChain:
			ilog.Infof("FoundChain: %v, %v", t, common.Base58Encode(t.Hash()))
			return errTxDup
		case txpool.NotFound:
			err := t.VerifySelf()
			if err != nil {
				return err
			}

		}
	}
	v := verifier.Verifier{}
	return v.Verify(blk, parent, witnessList, db, &verifier.Config{
		Mode:        0,
		Timeout:     genBlockTime,
		TxTimeLimit: common.MaxTxTimeLimit,
	})
}
