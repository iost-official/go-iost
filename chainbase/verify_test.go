package chainbase

import (
	"testing"
	"time"

	"github.com/iost-official/go-iost/v3/account"
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/core/tx"
	"github.com/iost-official/go-iost/v3/crypto"
	"github.com/smartystreets/goconvey/convey"
)

func TestVerifyBlock(t *testing.T) {
	convey.Convey("Test of verify block", t, func() {
		secKey := common.Sha3([]byte("secKey of id0"))
		account0, _ := account.NewKeyPair(secKey, crypto.Secp256k1)
		secKey = common.Sha3([]byte("sec of id1"))
		account1, _ := account.NewKeyPair(secKey, crypto.Secp256k1)
		secKey = common.Sha3([]byte("sec of id2"))
		account2, _ := account.NewKeyPair(secKey, crypto.Secp256k1)
		witnessList := []string{account0.ReadablePubkey(), account1.ReadablePubkey(), account2.ReadablePubkey()}
		rootTime := time.Now().UnixNano()
		rootBlk := &block.Block{
			Head: &block.BlockHead{
				Number:  1,
				Time:    rootTime,
				Witness: common.WitnessOfNanoSec(rootTime, witnessList),
			},
		}
		tx0 := &tx.Tx{
			Time: time.Now().UnixNano(),
			Actions: []*tx.Action{{
				Contract:   "contract1",
				ActionName: "actionname1",
				Data:       "{\"num\": 1, \"message\": \"contract1\"}",
			}},
			Signers: []string{account1.ReadablePubkey()},
		}
		rcpt0 := &tx.TxReceipt{
			TxHash: tx0.Hash(),
		}
		curTime := time.Now().UnixNano()
		hash := rootBlk.HeadHash()
		witness := common.WitnessOfNanoSec(curTime, witnessList)
		blk := &block.Block{
			Head: &block.BlockHead{
				Number:     2,
				ParentHash: hash,
				Time:       curTime,
				Witness:    common.WitnessOfNanoSec(curTime, witnessList),
			},
			Txs:      []*tx.Tx{},
			Receipts: []*tx.TxReceipt{},
		}
		blk.Head.TxMerkleHash = blk.CalculateTxMerkleHash()
		blk.Head.TxReceiptMerkleHash = blk.CalculateTxReceiptMerkleHash()
		info := blk.HeadHash()
		var sig *crypto.Signature
		if witness == account0.ReadablePubkey() {
			sig = account0.Sign(info)
		} else if witness == account1.ReadablePubkey() {
			sig = account1.Sign(info)
		} else {
			sig = account2.Sign(info)
		}
		blk.Sign = sig
		//convey.Convey("Normal (no txs)", func() {
		//      err := verifyBlock(blk, rootBlk, rootBlk, nil, nil)
		//      convey.So(err, convey.ShouldBeNil)
		//})

		convey.Convey("Normal (with txs)", func() {
			blk.Txs = append(blk.Txs, tx0)
			blk.Receipts = append(blk.Receipts, rcpt0)
			//Use mock
			//txPool, _ := txpool.NewTxPoolImpl()
			//db, _ := db.NewMVCCDB()
			//err := verifyBlock(blk, rootBlk, rootBlk, txPool, db)
			//convey.So(err, convey.ShouldBeNil)
		})
	})
}
