package pob

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/cverifier"
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
	errWitness     = errors.New("wrong witness")
	errSignature   = errors.New("wrong signature")
	errTxDup       = errors.New("duplicate tx")
	errDoubleTx    = errors.New("double tx in block")
	errTxSignature = errors.New("tx wrong signature")
	errHeadHash    = errors.New("wrong head hash")
	generateTxsNum = 0
)

func saveBlockHead(db db.MVCCDB, bh *block.BlockHead) error {
	bhJson, err := json.Marshal(bh)
	if err != nil {
		return fmt.Errorf("json fail: %v", err)
	}
	err = db.Put("snapshot", "blockHead", string(bhJson))
	if err != nil {
		return fmt.Errorf("state db put fail: %v", err)
	}
	return nil
}

func generateBlock(acc *account.KeyPair, txPool txpool.TxPool, db db.MVCCDB, limitTime time.Duration) (*block.Block, error) { // TODO 应传入account
	ilog.Debug("generate Block start")
	st := time.Now()
	pTx, head := txPool.PendingTx()
	topBlock := head.Block
	blk := block.Block{
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
	dropList, _, err := v.Gen(&blk, topBlock, db, pTx, &verifier.Config{
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
	err = saveBlockHead(db, blk.Head)
	if err != nil {
		return nil, err
	}
	db.Tag(string(blk.HeadHash()))
	metricsGeneratedBlockCount.Add(1, nil)
	generateTxsNum += len(blk.Txs)
	return &blk, nil
}

func verifyBasics(blk *block.Block, signature *crypto.Signature) error {
	signature.SetPubkey(account.DecodePubkey(blk.Head.Witness))
	hash := blk.HeadHash()
	if !signature.Verify(hash) {
		return errSignature
	}
	return nil
}

func verifyBlock(blk *block.Block, parent *block.Block, lib *block.Block, txPool txpool.TxPool, db db.MVCCDB, chain block.Chain, replay bool) error {
	err := cverifier.VerifyBlockHead(blk, parent, lib)
	if err != nil {
		return err
	}

	if replay == false && witnessOfNanoSec(blk.Head.Time) != blk.Head.Witness {
		ilog.Errorf("blk num: %v, time: %v, witness: %v, witness len: %v, witness list: %v",
			blk.Head.Number, blk.Head.Time, blk.Head.Witness, staticProperty.NumberOfWitnesses, staticProperty.WitnessList)
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
				return errTxSignature
			}

		}
		if t.IsDefer() {
			referredTx, err := chain.GetTx(t.ReferredTx)
			if err != nil {
				return fmt.Errorf("get referred tx error, %v", err)
			}
			err = t.VerifyDefer(referredTx)
			if err != nil {
				return err
			}
		}
	}
	v := verifier.Verifier{}
	return v.Verify(blk, parent, db, &verifier.Config{
		Mode:        0,
		Timeout:     time.Millisecond * 250,
		TxTimeLimit: time.Millisecond * 100,
	})
}

func updateWaterMark(node *blockcache.BlockCacheNode) {
	node.ConfirmUntil = staticProperty.Watermark[node.Head.Witness]
	if node.Head.Number >= staticProperty.Watermark[node.Head.Witness] {
		staticProperty.Watermark[node.Head.Witness] = node.Head.Number + 1
	}
}

func updateLib(node *blockcache.BlockCacheNode, bc blockcache.BlockCache) {
	confirmedNode := calculateConfirm(node, bc.LinkedRoot())
	if confirmedNode != nil {
		bc.Flush(confirmedNode)
		metricsConfirmedLength.Set(float64(confirmedNode.Head.Number+1), nil)
	}
}

func calculateConfirm(node *blockcache.BlockCacheNode, root *blockcache.BlockCacheNode) *blockcache.BlockCacheNode {
	confirmLimit := staticProperty.NumberOfWitnesses*2/3 + 1
	startNumber := node.Head.Number
	var confirmNum int64
	confirmUntilMap := make(map[int64]int64, startNumber-root.Head.Number)
	for node != root {
		if node.ConfirmUntil <= node.Head.Number {
			confirmNum++
			confirmUntilMap[node.ConfirmUntil]++
		}
		if confirmNum >= confirmLimit {
			return node
		}
		confirmNum -= confirmUntilMap[node.Head.Number]
		node = node.GetParent()
	}
	return nil
}
