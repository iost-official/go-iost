package block

import (
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewBlockChain(t *testing.T) {
	Convey("test TestNewBlockChain", t, func() {
		txDb := tx.TxDbInstance()
		So(txDb, ShouldNotBeNil)

		bc, err := Instance()
		Convey("New", func() {
			So(err, ShouldBeNil)
			So(bc.Length(), ShouldEqual, bc.Length())
		})
	})
}

func TestChainImpl(t *testing.T) {
	Convey("test Push", t, func() {
		txDb := tx.TxDbInstance()
		So(txDb, ShouldNotBeNil)

		bc, err := Instance()

		Convey("New", func() {
			So(err, ShouldBeNil)

		})

		tBlock := Block{Head: BlockHead{
			Version:    2,
			ParentHash: []byte("parent Hash"),
			TreeHash:   []byte("tree hash"),
			Info:       []byte("info "),
			Number:     int64(0),
			Witness:    "id2,id3,id5,id6",
			Signature:  []byte("Signatrue"),
			Time:       201222,
		}}

		err = state.PoolInstance()
		if err != nil {
			panic("state.PoolInstance error")
		}

		Convey("Push", func() {
			length := bc.Length()

			tBlock.Head.Number = int64(length)

			err = bc.Push(&tBlock)
			So(err, ShouldBeNil)
			So(bc.Length(), ShouldEqual, length+1)
		})

		Convey("GetBlockByNumber", func() {
			length := bc.Length()
			//取出刚存入的块
			tBlock.Head.Number = int64(length) - 1

			block := bc.GetBlockByNumber(length - 1)
			So(block, ShouldNotBeNil)
			So(block.Head.Version, ShouldEqual, 2)
			So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
			So(string(block.Head.TreeHash), ShouldEqual, string(tBlock.Head.TreeHash))
			So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
			So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
			So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
			So(string(block.Head.Signature), ShouldEqual, string(tBlock.Head.Signature))
			So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))

			block = bc.Top()
			So(block, ShouldNotBeNil)
			So(block.Head.Version, ShouldEqual, 2)
			So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
			So(string(block.Head.TreeHash), ShouldEqual, string(tBlock.Head.TreeHash))
			So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
			So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
			So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
			So(string(block.Head.Signature), ShouldEqual, string(tBlock.Head.Signature))
			So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))

			block = bc.GetBlockByHash(tBlock.Hash())
			So(block, ShouldNotBeNil)
			//fmt.Printf("###Top() block = %s\n", block)
			So(block.Head.Version, ShouldEqual, 2)
			So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
			So(string(block.Head.TreeHash), ShouldEqual, string(tBlock.Head.TreeHash))
			So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
			So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
			So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
			So(string(block.Head.Signature), ShouldEqual, string(tBlock.Head.Signature))
			So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))
		})
	})
}
