package cverifier

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	. "github.com/smartystreets/goconvey/convey"
)

func TestVerifyBlockHead(t *testing.T) {
	Convey("Test of verify block head", t, func() {
		parentBlk := &block.Block{
			Head: &block.BlockHead{
				Number: 3,
				Time:   common.GetCurrentTimestamp().Slot - 1,
			},
		}
		chainTop := &block.Block{
			Head: &block.BlockHead{
				Number: 1,
				Time:   common.GetCurrentTimestamp().Slot - 4,
			},
		}
		hash := parentBlk.HeadHash()
		blk := &block.Block{
			Head: &block.BlockHead{
				ParentHash: hash,
				Number:     4,
				Time:       common.GetCurrentTimestamp().Slot,
				TxsHash:    common.Sha3([]byte{}),
				MerkleHash: []byte{},
			},
		}
		convey.Convey("Pass", func() {
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Wrong time", func() {
			blk.Head.Time = common.GetCurrentTimestamp().Slot - 5
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, errOldBlk)
			blk.Head.Time = common.GetCurrentTimestamp().Slot + 2
			err = VerifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, errFutureBlk)
		})

		convey.Convey("Wrong parent", func() {
			blk.Head.ParentHash = []byte("fake hash")
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, errParentHash)
		})

		convey.Convey("Wrong number", func() {
			blk.Head.Number = 5
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, errNumber)
		})

		convey.Convey("Wrong tx hash", func() {
			tx0 := tx.NewTx(nil, nil, 1000, 1, 300, 0)
			blk.Txs = append(blk.Txs, tx0)
			blk.Head.TxsHash = blk.CalculateTxsHash()
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldBeNil)
			blk.Head.TxsHash = []byte("fake hash")
			err = VerifyBlockHead(blk, parentBlk, chainTop)
			convey.So(err, convey.ShouldEqual, errTxHash)
		})
	})
}
