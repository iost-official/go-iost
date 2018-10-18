package pob

import (
	"errors"
	"time"

	"fmt"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/verifier"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm"
)

var (
	errWitness     = errors.New("wrong witness")
	errSignature   = errors.New("wrong signature")
	errTxTooOld    = errors.New("tx too old")
	errTxDup       = errors.New("duplicate tx")
	errTxSignature = errors.New("tx wrong signature")
	errHeadHash    = errors.New("wrong head hash")
	txLimit        = 80000 //limit it to 2000
	txExecTime     = verifier.TxExecTimeLimit / 2
)

func generateBlock(account *account.Account, txPool txpool.TxPool, db db.MVCCDB) (*block.Block, error) {
	ilog.Info("[pob] generate Block start")
	limitTime := time.NewTimer(time.Second / 20)
	txIter, head := txPool.TxIterator()
	topBlock := head.Block
	blk := block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: topBlock.HeadHash(),
			Number:     topBlock.Head.Number + 1,
			Witness:    account.ID,
			Time:       time.Now().Unix() / common.SlotLength,
		},
		Txs:      []*tx.Tx{},
		Receipts: []*tx.TxReceipt{},
	}
	db.Checkout(string(topBlock.HeadHash()))
	engine := vm.NewEngine(blk.Head, db)

	// call vote
	if blk.Head.Number%common.VoteInterval == 0 {
		ilog.Info("vote start")
		act := tx.NewAction("iost.vote", "Stat", fmt.Sprintf(`[]`))
		trx := tx.NewTx([]*tx.Action{&act}, nil, 100000000, 0, 0)

		trx, err := tx.SignTx(trx, staticProperty.account)
		if err != nil {
			ilog.Errorf("fail to signTx, err:%v", err)
		}
		receipt, err := engine.Exec(trx, txExecTime)
		if err != nil {
			ilog.Errorf("fail to exec trx, err:%v", err)
		}
		if receipt.Status.Code != tx.Success {
			ilog.Errorf("status code: %v", receipt.Status.Code)
		}
		blk.Txs = append(blk.Txs, trx)
		blk.Receipts = append(blk.Receipts, receipt)
	}
	t, ok := txIter.Next()
	delList := []*tx.Tx{}
L:
	for ok {
		select {
		case <-limitTime.C:
			ilog.Info("time up")
			break L
		default:
			if !txPool.TxTimeOut(t) {
				if receipt, err := engine.Exec(t, txExecTime); err == nil {
					blk.Txs = append(blk.Txs, t)
					blk.Receipts = append(blk.Receipts, receipt)
				} else {
					ilog.Errorf("exec tx failed. err=%v, receipt=%v", err, receipt)
					delList = append(delList, t)
				}
			} else {
				delList = append(delList, t)
			}
			if len(blk.Txs) >= txLimit {
				break L
			}
			t, ok = txIter.Next()
		}
	}

	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	err := blk.CalculateHeadHash()
	if err != nil {
		return nil, err
	}
	blk.Sign = account.Sign(blk.HeadHash())
	db.Tag(string(blk.HeadHash()))

	metricsGeneratedBlockCount.Add(1, nil)
	metricsTxSize.Set(float64(len(blk.Txs)), nil)
	go txPool.DelTxList(delList)
	ilog.Infof("[pob] generate Block end, number: %v, hash = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()))
	ilog.Infof("[pob] block size: %v", len(blk.Txs))
	return &blk, nil
}

func verifyBasics(head *block.BlockHead, signature *crypto.Signature) error {

	signature.SetPubkey(account.GetPubkeyByID(head.Witness))
	hash, err := head.Hash()
	if err != nil {
		return errHeadHash
	}
	if !signature.Verify(hash) {
		return errSignature
	}
	return nil
}

//nolint
func verifyBlock(blk *block.Block, parent *block.Block, lib *block.Block, txPool txpool.TxPool, db db.MVCCDB) error {
	ilog.Infof("[pob] verifyBlockHead start, number: %d, hash = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()))
	err := verifier.VerifyBlockHead(blk, parent, lib)
	if err != nil {
		return err
	}

	if witnessOfSlot(blk.Head.Time) != blk.Head.Witness {
		ilog.Errorf("blk num: %v, time: %v, witness: %v, witness len: %v, witness list: %v",
			blk.Head.Number, blk.Head.Time, blk.Head.Witness, staticProperty.NumberOfWitnesses, staticProperty.WitnessList)
		return errWitness
	}

	// if it's vote block, check for votes
	if blk.Head.Number%common.VoteInterval == 0 {
		if len(blk.Txs) == 0 || len(blk.Txs[0].Actions) == 0 ||
			blk.Txs[0].Actions[0].Contract != "iost.vote" ||
			blk.Txs[0].Actions[0].ActionName != "Stat" ||
			blk.Txs[0].Actions[0].Data != "[]" {

			return errors.New("blk did not vote")
		}

		if blk.Receipts[0].Status.Code != tx.Success {
			return fmt.Errorf("vote was incorrect, status:%v", blk.Receipts[0].Status)
		}
	}
	// check txs
	ilog.Infof("[pob] verify tx in txpool start, number: %d, hash = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()))
	var notFoundPending int64

	for _, tx := range blk.Txs {
		exist := txPool.ExistTxs(tx.Hash(), parent)
		switch exist {
		case txpool.FoundChain:
			return errTxDup
		case txpool.NotFound:
			notFoundPending += 1
			err := tx.VerifySelf()
			if err != nil {
				return errTxSignature
			}
		case txpool.FoundPending:
		}
		if blk.Head.Time*common.SlotLength-tx.Time/1e9 > txpool.Expiration {
			return errTxTooOld
		}
	}
	ilog.Infof("[pob] all tx: %v, tx not found in pending: %v", len(blk.Txs), notFoundPending)
	ilog.Infof("[pob] verify tx in txpool end, number: %d, hash = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()))
	return verifier.VerifyBlockWithVM(blk, db)
}

func updateWaterMark(node *blockcache.BlockCacheNode) {
	node.ConfirmUntil = staticProperty.Watermark[node.Witness]
	if node.Number >= staticProperty.Watermark[node.Witness] {
		staticProperty.Watermark[node.Witness] = node.Number + 1
	}
}

func updateLib(node *blockcache.BlockCacheNode, bc blockcache.BlockCache) {
	ilog.Infof("[pob] calculate start, number: %d, hash = %v", node.Number, common.Base58Encode(node.Block.HeadHash()))
	confirmedNode := calculateConfirm(node, bc.LinkedRoot())
	ilog.Infof("[pob] calculate end, number: %d, hash = %v", node.Number, common.Base58Encode(node.Block.HeadHash()))
	if confirmedNode != nil {
		ilog.Infof("[pob] flush start, number: %d, hash = %v", node.Number, common.Base58Encode(node.Block.HeadHash()))
		bc.Flush(confirmedNode)
		ilog.Infof("[pob] flush end, number: %d, hash = %v", node.Number, common.Base58Encode(node.Block.HeadHash()))
		metricsConfirmedLength.Set(float64(confirmedNode.Number+1), nil)
	}
}

func calculateConfirm(node *blockcache.BlockCacheNode, root *blockcache.BlockCacheNode) *blockcache.BlockCacheNode {
	confirmLimit := staticProperty.NumberOfWitnesses*2/3 + 1
	startNumber := node.Number
	var confirmNum int64
	confirmUntilMap := make(map[int64]int64, startNumber-root.Number)
	for node != root {
		if node.ConfirmUntil <= node.Number {
			confirmNum++
			confirmUntilMap[node.ConfirmUntil]++
		}
		if confirmNum >= confirmLimit {
			return node
		}
		confirmNum -= confirmUntilMap[node.Number]
		node = node.Parent
	}
	return nil
}
