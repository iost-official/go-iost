package block

import (
	"bytes"
	"testing"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBlockHeadSerialize(t *testing.T) {
	Convey("Test of block head encode and decode", t, func() {
		head := &BlockHead{
			Number:     1,
			Witness:    "id1",
			ParentHash: []byte("parent"),
		}
		h := head.Encode()
		var head1 BlockHead
		err := head1.Decode(h)
		So(err, ShouldBeNil)

		So(bytes.Equal(head.ParentHash, head1.ParentHash), ShouldBeTrue)
		So(head1.Number == head.Number, ShouldBeTrue)
		So(head1.Witness == head.Witness, ShouldBeTrue)
	})
}

func TestBlockSerialize(t *testing.T) {
	Convey("Test of block encode and decode", t, func() {
		blk := &Block{
			Head: BlockHead{
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
		b := blk.Encode()
		var blk1 Block
		err := blk1.Decode(b)
		So(err, ShouldBeNil)

		So(bytes.Equal(blk1.Head.ParentHash, blk.Head.ParentHash), ShouldBeTrue)
		So(len(blk1.Txs) == len(blk.Txs), ShouldBeTrue)
		So(len(blk1.Receipts) == len(blk.Receipts), ShouldBeTrue)
		So(bytes.Equal(blk1.Receipts[0].TxHash, tx0.Hash()), ShouldBeTrue)
	})
}
