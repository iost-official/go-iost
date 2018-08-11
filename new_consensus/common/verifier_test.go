package consensus_common

import (
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

func TestVerifyBlockHead(t *testing.T) {
	Convey("Test of verify block head", t, func() {
		parentBlk := &block.Block{
			Head: block.BlockHead{
				Number: 3,
				Time:   GetCurrentTimestamp().Slot - 1,
			},
		}
		chainTop := &block.Block{
			Head: block.BlockHead{
				Number: 1,
				Time:   GetCurrentTimestamp().Slot - 4,
			},
		}
		hash, _ := parentBlk.HeadHash()
		blk := &block.Block{
			Head: block.BlockHead{
				ParentHash: hash,
				Number:     4,
				Time:       GetCurrentTimestamp().Slot,
				TxsHash:    common.Sha256([]byte{}),
				MerkleHash: []byte{},
			},
		}
		Convey("Pass", func() {
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			So(err, ShouldBeNil)
		})

		Convey("Wrong time", func() {
			blk.Head.Time = GetCurrentTimestamp().Slot - 5
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			So(err, ShouldEqual, ErrOldBlk)
			blk.Head.Time = GetCurrentTimestamp().Slot + 2
			err = VerifyBlockHead(blk, parentBlk, chainTop)
			So(err, ShouldEqual, ErrFutureBlk)
		})

		Convey("Wrong parent", func() {
			blk.Head.ParentHash = []byte("fake hash")
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			So(err, ShouldEqual, ErrParentHash)
		})

		Convey("Wrong number", func() {
			blk.Head.Number = 5
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			So(err, ShouldEqual, ErrNumber)
		})

		Convey("Wrong tx hash", func() {
			tx0 := tx.NewTx(nil, nil, 1000, 1, 300)
			blk.Txs = append(blk.Txs, &tx0)
			blk.Head.TxsHash, _ = blk.CalculateTxsHash()
			err := VerifyBlockHead(blk, parentBlk, chainTop)
			So(err, ShouldBeNil)
			blk.Head.TxsHash = []byte("fake hash")
			err = VerifyBlockHead(blk, parentBlk, chainTop)
			So(err, ShouldEqual, ErrTxHash)
		})
	})
}
