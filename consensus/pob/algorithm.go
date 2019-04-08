package pob

import (
	"fmt"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/verifier"
)

func (p *PoB) generateBlock(limitTime time.Duration) (*block.Block, error) {
	ilog.Debug("generate Block start")

	st := time.Now()
	pTx, head := p.txPool.PendingTx()
	witnessList := head.Active()
	if common.WitnessOfNanoSec(st.UnixNano(), witnessList) != p.account.ReadablePubkey() {
		return nil, fmt.Errorf("Now time %v exceeding the slot of witness %v", st.UnixNano(), p.account.ReadablePubkey())
	}
	blk := &block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: head.HeadHash(),
			Info:       make([]byte, 0),
			Number:     head.Head.Number + 1,
			Witness:    p.account.ReadablePubkey(),
			Time:       time.Now().UnixNano(),
		},
		Txs:      []*tx.Tx{},
		Receipts: []*tx.TxReceipt{},
	}
	p.produceDB.Checkout(string(head.HeadHash()))

	// call vote
	v := verifier.Verifier{}
	t1 := time.Now()
	// TODO: stateDb and block head is consisdent, pTx may be inconsisdent.
	dropList, _, err := v.Gen(blk, head.Block, &head.WitnessList, p.produceDB, pTx, &verifier.Config{
		Mode:        0,
		Timeout:     limitTime - time.Now().Sub(st),
		TxTimeLimit: common.MaxTxTimeLimit,
	})
	t2 := time.Since(t1)
	if len(blk.Txs) != 0 {
		ilog.Debugf("time spent per tx: %v", t2.Nanoseconds()/int64(len(blk.Txs)))
	}
	if err != nil {
		go p.delTxList(dropList)
		ilog.Errorf("Gen is err: %v", err)
		return nil, err
	}
	blk.Head.TxMerkleHash = blk.CalculateTxMerkleHash()
	blk.Head.TxReceiptMerkleHash = blk.CalculateTxReceiptMerkleHash()
	err = blk.CalculateHeadHash()
	if err != nil {
		return nil, err
	}
	blk.Sign = p.account.Sign(blk.HeadHash())
	p.produceDB.Commit(string(blk.HeadHash()))
	return blk, nil
}

func (p *PoB) delTxList(delList []*tx.Tx) {
	for _, t := range delList {
		p.txPool.DelTx(t.Hash())
	}
}
