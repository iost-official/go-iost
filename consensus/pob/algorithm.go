package pob

import (
	"errors"
	"time"

	"fmt"
	"strings"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus/verifier"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/core/txpool"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm"
)

var (
	errWitness     = errors.New("wrong witness")
	errSignature   = errors.New("wrong signature")
	errSlot        = errors.New("witness slot duplicate")
	errTxTooOld    = errors.New("tx too old")
	errTxDup       = errors.New("duplicate tx")
	errTxSignature = errors.New("tx wrong signature")
	errHeadHash    = errors.New("wrong head hash")
)

func generateBlock(account *account.Account, txPool txpool.TxPool, db db.MVCCDB) (*block.Block, error) {
	ilog.Info("generate Block start")
	limitTime := time.NewTimer(common.SlotLength / 3 * time.Second)
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
		receipt, err := engine.Exec(trx)
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

	var vmExecTime, iterTime, i, j int64
L:
	for ok {
		select {
		case <-limitTime.C:
			ilog.Info("time up")
			break L
		default:
			i++
			step1 := time.Now()
			if !txPool.TxTimeOut(t) {
				j++
				if receipt, err := engine.Exec(t); err == nil {
					blk.Txs = append(blk.Txs, t)
					blk.Receipts = append(blk.Receipts, receipt)
					ilog.Debug(err, receipt)
				} else {
					ilog.Errorf("exec tx failed. err=%v, receipt=%v", err, receipt)
					txPool.DelTx(t.Hash())
				}
			}
			step2 := time.Now()
			t, ok = txIter.Next()
			step3 := time.Now()
			vmExecTime += step2.Sub(step1).Nanoseconds()
			iterTime += step3.Sub(step2).Nanoseconds()
		}
	}
	if i > 0 && j > 0 {
		metricsVMTime.Set(float64(vmExecTime), nil)
		metricsVMAvgTime.Set(float64(vmExecTime/j), nil)
		metricsIterTime.Set(float64(iterTime), nil)
		metricsIterAvgTime.Set(float64(iterTime/j), nil)
		metricsNonTimeOutTxSize.Set(float64(j), nil)
		metricsAllTxSize.Set(float64(i), nil)
		ilog.Infof("tx in blk:%d, iter:%d, vmExecTime:%d, vmAvgTime:%d, iterTime:%d, iterAvgTime:%d",
			len(blk.Txs), i, vmExecTime, vmExecTime/j, iterTime, iterTime/j)
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
	err := verifier.VerifyBlockHead(blk, parent, lib)
	if err != nil {
		return err
	}

	if witnessOfSlot(blk.Head.Time) != blk.Head.Witness {
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
		exist, _ := txPool.ExistTxs(tx.Hash(), parent)
		if exist == txpool.FoundChain {
			return errTxDup
		} else if exist != txpool.FoundPending {
			if err := tx.VerifySelf(); err != nil {
				return errTxSignature
			}
		}
		if blk.Head.Time*common.SlotLength-tx.Time/1e9 > 60 {
			return errTxTooOld
		}
	}
	return verifier.VerifyBlockWithVM(blk, db)
}

func updateWaterMark(node *blockcache.BlockCacheNode) {
	node.ConfirmUntil = staticProperty.Watermark[node.Witness]
	if node.Number >= staticProperty.Watermark[node.Witness] {
		staticProperty.Watermark[node.Witness] = node.Number + 1
	}
}

func updateLib(node *blockcache.BlockCacheNode, bc blockcache.BlockCache) {
	confirmedNode := calculateConfirm(node, bc.LinkedRoot())
	// bc.Flush(node) // debug do not delete this
	if confirmedNode != nil {
		bc.Flush(confirmedNode)
		metricsConfirmedLength.Set(float64(confirmedNode.Number+1), nil)
	}
}

func calculateConfirm(node *blockcache.BlockCacheNode, root *blockcache.BlockCacheNode) *blockcache.BlockCacheNode {
	confirmLimit := staticProperty.NumberOfWitnesses*2/3 + 1
	startNumber := node.Number
	var confirmNum int64
	confirmUntilMap := make(map[int64]int64, startNumber-root.Number)
	var index int64
	for node != root {
		if node.ConfirmUntil <= node.Number {
			confirmNum++
			confirmUntilMap[startNumber-node.ConfirmUntil]++
		}
		if confirmNum >= confirmLimit {
			return node
		}
		confirmNum -= confirmUntilMap[index]
		node = node.Parent
		index++
	}
	return nil
}
