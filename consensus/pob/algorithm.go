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
	generateTxsNum = 0
)

func generateBlock(acc *account.KeyPair, txPool txpool.TxPool, db db.MVCCDB, limitTime time.Duration) (*block.Block, error) { // TODO 应传入account
	ilog.Info("[pob]generate Block start")
	st := time.Now()
	txIter, head := txPool.TxIterator()
	topBlock := head.Block
	blk := block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: topBlock.HeadHash(),
			Number:     topBlock.Head.Number + 1,
			Witness:    acc.ID,
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
		trx := tx.NewTx([]*tx.Action{act}, nil, 100000000, 0, 0, 0)

		trx, err := tx.SignTx(trx, staticProperty.account.ID, []*account.KeyPair{staticProperty.account})
		if err != nil {
			ilog.Errorf("fail to signTx, err:%v", err)
		}
		receipt, err := v.Exec(blk.Head, db, trx, time.Millisecond*100)
		if err != nil {
			ilog.Errorf("fail to exec trx, err:%v", err)
		}
		if receipt.Status.Code != tx.Success { //rt is nil
			ilog.Errorf("status code: %v", receipt.Status.Code)
		}
		blk.Txs = append(blk.Txs, trx)
		blk.Receipts = append(blk.Receipts, receipt)
	}
	t1 := time.Now()
	dropList, _, err := v.Gen(&blk, db, txIter, &verifier.Config{
		Mode:        0,
		Timeout:     limitTime - time.Now().Sub(st),
		TxTimeLimit: time.Millisecond * 100,
	})
	t2 := time.Since(t1)
	ilog.Info("time spent:", t2)
	if len(blk.Txs) != 0 {
		ilog.Info("time spent per tx:", t2.Nanoseconds()/int64(len(blk.Txs)))
	}
	if err != nil {
		go txPool.DelTxList(dropList)
	}
	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	err = blk.CalculateHeadHash()
	if err != nil {
		return nil, err
	}
	blk.Sign = acc.Sign(blk.HeadHash())
	db.Tag(string(blk.HeadHash()))
	ilog.Infof("generate block txs num: %v, %v, %v", len(blk.Txs), blk.Head.Number, blk.Head.Witness)
	metricsGeneratedBlockCount.Add(1, nil)
	generateTxsNum += len(blk.Txs)
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
	// check txs
	ilog.Infof("[pob] start to verify block if foundchain, number: %v, hash = %v, witness = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()), blk.Head.Witness[4:6])
	for _, tx := range blk.Txs {
		exist := txPool.ExistTxs(tx.Hash(), parent)
		switch exist {
		case txpool.FoundChain:
			ilog.Infof("FoundChain: %v, %v", tx, common.Base58Encode(tx.Hash()))
			return errTxDup
		case txpool.NotFound:
			err := tx.VerifySelf()
			if err != nil {
				return errTxSignature
			}
		}
		if blk.Head.Time*common.SlotLength-tx.Time/1e9 > txpool.Expiration {
			return errTxTooOld
		}
	}
	v := verifier.Verifier{}
	ilog.Infof("[pob] start to verify block in vm, number: %v, hash = %v, witness = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()), blk.Head.Witness[4:6])
	defer ilog.Infof("[pob] end of verify block in vm, number: %v, hash = %v, witness = %v", blk.Head.Number, common.Base58Encode(blk.HeadHash()), blk.Head.Witness[4:6])
	return v.Verify(blk, db, &verifier.Config{
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
		ilog.Infof("[pob] flush start, number: %d, hash = %v", node.Head.Number, common.Base58Encode(node.Block.HeadHash()))
		bc.Flush(confirmedNode)
		ilog.Infof("[pob] flush end, number: %d, hash = %v", node.Head.Number, common.Base58Encode(node.Block.HeadHash()))
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
