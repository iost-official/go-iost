package block

import (
	"bytes"
	"testing"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/smartystreets/goconvey/convey"
)

func TestBlockHeadSerialize(t *testing.T) {
	convey.Convey("Test of block head encode and decode", t, func() {
		head := BlockHead{
			Number:     1,
			Witness:    "id1",
			ParentHash: []byte("parent"),
		}
		hash, err := head.Encode()
		convey.So(err, convey.ShouldBeNil)
		var head_read BlockHead
		err = head_read.Decode(hash)
		convey.So(err, convey.ShouldBeNil)
		convey.So(bytes.Equal(head.ParentHash, head_read.ParentHash), convey.ShouldBeTrue)
		convey.So(head_read.Number == head.Number, convey.ShouldBeTrue)
		convey.So(head_read.Witness == head.Witness, convey.ShouldBeTrue)
	})
}

func TestBlockSerialize(t *testing.T) {
	convey.Convey("Test of block encode and decode", t, func() {
		blk := Block{
			Head: &BlockHead{
				Number:     1,
				Witness:    "id1",
				ParentHash: []byte("parent"),
			},
		}
		a1, _ := account.NewAccount(nil)
		tx0 := tx.Tx{
			Time: 1,
			Actions: []tx.Action{{
				Contract:   "contract1",
				ActionName: "actionname1",
				Data:       "{\"num\": 1, \"message\": \"contract1\"}",
			}},
			Signers: [][]byte{a1.Pubkey},
		}
		blk.Txs = append(blk.Txs, &tx0)
		receipt := tx.TxReceipt{
			TxHash:   tx0.Hash(),
			GasUsage: 10,
			Status: tx.Status{
				Code:    tx.Success,
				Message: "run success",
			},
		}
		blk.Receipts = append(blk.Receipts, &receipt)
		blkByte, err := blk.Encode()
		blkRead := Block{}
		err = blkRead.Decode(blkByte)
		convey.So(err, convey.ShouldBeNil)
		convey.So(bytes.Equal(blkRead.Head.ParentHash, blk.Head.ParentHash), convey.ShouldBeTrue)
		convey.So(len(blkRead.Txs) == len(blk.Txs), convey.ShouldBeTrue)
		convey.So(len(blkRead.Receipts) == len(blk.Receipts), convey.ShouldBeTrue)
		convey.So(bytes.Equal(blkRead.Receipts[0].TxHash, tx0.Hash()), convey.ShouldBeTrue)
		convey.So(err, convey.ShouldBeNil)

	})
}
