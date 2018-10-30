package pob

import (
	"errors"
	"time"

	"fmt"
	"strings"

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
	errTxTooOld    = errors.New("tx too old")
	errTxDup       = errors.New("duplicate tx")
	errTxSignature = errors.New("tx wrong signature")
	errHeadHash    = errors.New("wrong head hash")
	//txLimit        = 2000 //limit it to 2000
	//txExecTime     = cverifier.TxExecTimeLimit / 2
)

func generateBlock(account *account.KeyPair, txPool txpool.TxPool, db db.MVCCDB) (*block.Block, error) { // TODO 应传入account

	ilog.Info("generate Block start")
	st := time.Now()
	limitTime := common.SlotLength / 3 * time.Second
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

	// call vote
	v := verifier.Verifier{}
	if blk.Head.Number%common.VoteInterval == 0 {
		ilog.Info("vote start")
		act := tx.NewAction("iost.vote", "Stat", fmt.Sprintf(`[]`))
		trx := tx.NewTx([]*tx.Action{&act}, nil, 100000000, 0, 0)

		trx, err := tx.SignTx(trx, staticProperty.account)
		if err != nil {
			ilog.Errorf("fail to signTx, err:%v", err)
		}
		receipt, err := v.Exec(blk.Head, db, trx, time.Millisecond*100)
		if err != nil {
			ilog.Errorf("fail to exec trx, err:%v", err)
		}
		if receipt.Status.Code != tx.Success {
			ilog.Errorf("status code: %v", receipt.Status.Code)
		}
		blk.Txs = append(blk.Txs, trx)
		blk.Receipts = append(blk.Receipts, receipt)
	}
	dropList, _, err := v.Gen(&blk, db, txIter, &verifier.Config{
		Mode:        0,
		Timeout:     limitTime - st.Sub(time.Now()),
		TxTimeLimit: time.Millisecond * 100,
	})
	if err != nil {
		go txPool.DelTxList(dropList)
	}
	//t, ok := txIter.Next()
	//	var vmExecTime, iterTime, i, j int64
	//L:
	//	for ok {
	//		select {
	//		case <-limitTime.C:
	//			ilog.Info("time up")
	//			break L
	//		default:
	//			i++
	//			step1 := time.Now()
	//			if !txPool.TxTimeOut(t) {
	//				j++
	//				if receipt, err := engine.Exec(t, txExecTime); err == nil {
	//					blk.Txs = append(blk.Txs, t)
	//					blk.Receipts = append(blk.Receipts, receipt)
	//				} else {
	//					ilog.Errorf("exec tx failed. err=%v, receipt=%v", err, receipt)
	//					delList = append(delList, t)
	//				}
	//			} else {
	//				delList = append(delList, t)
	//			}
	//			if len(blk.Txs) >= txLimit {
	//				break L
	//			}
	//			step2 := time.Now()
	//			t, ok = txIter.Next()
	//			step3 := time.Now()
	//			vmExecTime += step2.Sub(step1).Nanoseconds()
	//			iterTime += step3.Sub(step2).Nanoseconds()
	//		}
	//	}

	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	err = blk.CalculateHeadHash()
	if err != nil {
		return nil, err
	}
	blk.Sign = account.Sign(blk.HeadHash())
	db.Tag(string(blk.HeadHash()))

	metricsGeneratedBlockCount.Add(1, nil)
	metricsTxSize.Set(float64(len(blk.Txs)), nil)
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

func verifyBlock(blk *block.Block, parent *block.Block, lib *block.Block, txPool txpool.TxPool, db db.MVCCDB) error {
	err := cverifier.VerifyBlockHead(blk, parent, lib)
	if err != nil {
		return err
	}

	if witnessOfSlot(blk.Head.Time) != blk.Head.Witness {
		ilog.Errorf("blk num: %v, time: %v, witness: %v, witness len: %v, witness list: %v",
			blk.Head.Number, blk.Head.Time, blk.Head.Witness, staticProperty.NumberOfWitnesses, staticProperty.WitnessList)
		return errWitness
	}

	// check vote
	if blk.Head.Number%common.VoteInterval == 0 {
		if len(blk.Txs) == 0 || strings.Compare(blk.Txs[0].Actions[0].Contract, "iost.vote") != 0 ||
			strings.Compare(blk.Txs[0].Actions[0].ActionName, "Stat") != 0 ||
			strings.Compare(blk.Txs[0].Actions[0].Data, fmt.Sprintf(`[]`)) != 0 {

			return errors.New("blk did not vote")
		}

		if blk.Receipts[0].Status.Code != tx.Success {
			return fmt.Errorf("vote was incorrect, status:%v", blk.Receipts[0].Status)
		}
	}

	for _, tx := range blk.Txs {
		exist := txPool.ExistTxs(tx.Hash(), parent)
		if exist == txpool.FoundChain {
			return errTxDup
		} else if exist != txpool.FoundPending {
			if err := tx.VerifySelf(); err != nil {
				return errTxSignature
			}
		}
		if blk.Head.Time*common.SlotLength-tx.Time/1e9 > txpool.Expiration {
			return errTxTooOld
		}
	}
	v := verifier.Verifier{}
	return v.Verify(blk, db, &verifier.Config{
		Mode:        0,
		Timeout:     common.SlotLength / 3 * time.Second,
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
		node = node.Parent
	}
	return nil
}
