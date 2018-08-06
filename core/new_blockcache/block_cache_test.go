package blockcache

import (
	"testing"
//	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/mocks"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	. "github.com/smartystreets/goconvey/convey"
)

func genBlock(fa *block.Block,wit string,num uint64) *block.Block {
	return &block.Block{
		Head: block.BlockHead{
			ParentHash: fa.HeadHash(),
			Witness: wit,
			Number: int64(num),
		},
	}
}
func TestBlockCache(t *testing.T) {
	ctl := gomock.NewController(t)
	b0 := &block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: []byte("nothing"),
			Witness:    "w0",
			Number: 0,
	},
}

	b1 := genBlock(b0,"w1",1)
	b2 := genBlock(b1,"w2",2)
	b2a := genBlock(b1,"w3",3)
	b3 := genBlock(b2,"w4",4)
	b4 := genBlock(b2a,"w5",5)

	base := core_mock.NewMockChain(ctl)
	base.EXPECT().Top().AnyTimes().Return(b0)
	block.BChain=base
	Convey("Test of Block Cache", t, func() {
		Convey("Add:", func() {
			bc := NewBlockCache(nil)
			//fmt.Printf("Leaf:%+v\n",bc.Leaf)
			b1node,_:=bc.Add(b1)
			//fmt.Printf("Leaf:%+v\n",bc.Leaf)
			bc.Add(b2)
			_,err := bc.Add(b2)
			So(err, ShouldEqual, ErrDup)
			_,err = bc.Add(b4)
			//fmt.Printf("Leaf:%+v\n",bc.Leaf)
			So(err, ShouldEqual, ErrNotFound)
			_,err = bc.Add(b4)
			So(err, ShouldEqual, ErrDup)
			bc.Del(b1node)
			//fmt.Printf("after Del Leaf:%+v\n",bc.Leaf)
			b1node,_=bc.Add(b1)
			//fmt.Printf("Leaf:%+v\n",bc.Leaf)
			//bc.Draw()
			So(bc.Head,ShouldEqual,b1node)
			bc.Add(b3)
			So(bc.Head,ShouldEqual,b1node)

})

		Convey("Longest chain", func() {
/*
			bc := NewBlockCache()

			bc.Add(&b1)
			b2node,_:=bc.Add(&b2)
			bc.Add(&b2a)
			So(bc.Head, ShouldEqual,b2node)
			b4node,_:=bc.Add(&b4)
			So(bc.Head, ShouldEqual,b4node)
*/
		})

	})
}

