package block

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"fmt"
				)

func TestNewBlockChain(t *testing.T) {
	Convey("test TestNewBlockChain", t, func() {
		bc, err := Instance()
		So(err, ShouldBeNil)
		So(bc.Length(), ShouldEqual, bc.Length())
		fmt.Println(bc.Length())
	})
}


func TestChainImpl(t *testing.T) {
	Convey("test Push", t, func() {
		bc, err := Instance()
		So(err, ShouldBeNil)
		tBlock := Block{Head: BlockHead{
			Version:    2,
			ParentHash: []byte("parent Hash"),
			TxsHash:    []byte("tree hash"),
			Info:       []byte("info "),
			Number:     int64(0),
			Witness:    "id2,id3,id5,id6",
			Signature:  []byte("Signatrue"),
			Time:       201222,
		}}
		//test Push
		length := bc.Length()
		fmt.Println("length:", length)
		tBlock.Head.Number = int64(length)
		err = bc.Push(&tBlock)
		So(err, ShouldBeNil)
		So(bc.Length(), ShouldEqual, length+1)

		//test GetBlockByNumber

		block, err := bc.GetBlockByNumber(bc.Length() - 1)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
		So(block.Head.Version, ShouldEqual, 2)
		So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
		So(string(block.Head.TxsHash), ShouldEqual, string(tBlock.Head.TxsHash))
		So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
		So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
		So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
		So(string(block.Head.Signature), ShouldEqual, string(tBlock.Head.Signature))
		So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))

		block, err = bc.Top()
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
		So(block.Head.Version, ShouldEqual, 2)
		So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
		So(string(block.Head.TxsHash), ShouldEqual, string(tBlock.Head.TxsHash))
		So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
		So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
		So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
		So(string(block.Head.Signature), ShouldEqual, string(tBlock.Head.Signature))
		So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))

		HeadHash, err := tBlock.HeadHash()
		block, err = bc.GetBlockByHash(HeadHash)
		So(err, ShouldBeNil)
		So(block, ShouldNotBeNil)
		So(block.Head.Version, ShouldEqual, 2)
		So(string(block.Head.ParentHash), ShouldEqual, string(tBlock.Head.ParentHash))
		So(string(block.Head.TxsHash), ShouldEqual, string(tBlock.Head.TxsHash))
		So(string(block.Head.Info), ShouldEqual, string(tBlock.Head.Info))
		So(block.Head.Number, ShouldEqual, tBlock.Head.Number)
		So(string(block.Head.Witness), ShouldEqual, string(tBlock.Head.Witness))
		So(string(block.Head.Signature), ShouldEqual, string(tBlock.Head.Signature))
		So(string(block.Head.Time), ShouldEqual, string(tBlock.Head.Time))
	})
}

