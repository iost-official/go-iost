package block

import (
	"testing"
	"github.com/smartystreets/goconvey/convey"
	"bytes"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/account"
)

func TestBlockHeadSerialize(t *testing.T) {
	convey.Convey("Test of block head encode and decode", t, func() {
		head := BlockHead{
			Number: 1,
			Witness: "id1",
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
		block := Block{
			Head: BlockHead{
				Number: 1,
				Witness: "id1",
				ParentHash: []byte("parent"),
			},
		}
		a1, _ := account.NewAccount(nil)
		tx0 := tx.Tx{
			Time: 1,
			Actions:[]tx.Action{{
				Contract:"contract1",
				ActionName:"actionname1",
				Data:"{\"num\": 1, \"message\": \"contract1\"}",
			}},
			Signers:[][]byte{a1.Pubkey},
		}
		block.Txs = append(block.Txs, tx0)
		receipt := tx.TxReceipt{
			TxHash: tx0.Hash(),
			GasUsage: 10,
			Status: tx.Status{
				Code: tx.Success,
				Message: "run success",
			},
		}
		block.Receipts = append(block.Receipts, receipt)
		blockByte, err := block.Encode()
		convey.So(err, convey.ShouldBeNil)
		var blockRead Block
		err = blockRead.Decode(blockByte)
		convey.So(err, convey.ShouldBeNil)
		convey.So(bytes.Equal(block.Head.ParentHash, blockRead.Head.ParentHash), convey.ShouldBeTrue)
		convey.So(len(block.Txs) == len(blockRead.Txs), convey.ShouldBeTrue)
		convey.So(len(block.Receipts) == len(blockRead.Receipts), convey.ShouldBeTrue)
		convey.So(bytes.Equal(blockRead.Receipts[0].TxHash, tx0.Hash()), convey.ShouldBeTrue)
	})
}