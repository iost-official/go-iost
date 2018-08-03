package blockcache

import (
	"testing"

	"errors"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/mocks"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/db/mocks"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	"github.com/iost-official/Go-IOS-Protocol/vm/lua"

	"github.com/iost-official/Go-IOS-Protocol/core/block"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBlockCache(t *testing.T) {
	ctl := gomock.NewController(t)

	main := lua.NewMethod(vm.Public, "main", 0, 1)
	code := `function main()
						Put("hello", "world")
						return "success"
					end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

	b0 := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: []byte("nothing"),
			Witness:    "w0",
		},
		Content: []tx.Tx{tx.NewTx(0, &lc)},
	}

	b1 := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: b0.HeadHash(),
			Witness:    "w1",
		},
		Content: []tx.Tx{tx.NewTx(1, &lc)},
	}

	b2 := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: b1.HeadHash(),
			Witness:    "w2",
		},
		Content: []tx.Tx{tx.NewTx(2, &lc)},
	}

	b2a := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: b1.HeadHash(),
			Witness:    "w3",
		},
		Content: []tx.Tx{tx.NewTx(-2, &lc)},
	}

	b3 := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: b2.HeadHash(),
			Witness:    "w1",
		},
		Content: []tx.Tx{tx.NewTx(3, &lc)},
	}

	b4 := block.Block{
		Head: block.BlockHead{
			Version:    0,
			ParentHash: b2a.HeadHash(),
			Witness:    "w2",
		},
		Content: []tx.Tx{tx.NewTx(4, &lc)},
	}

	base := core_mock.NewMockChain(ctl)
	base.EXPECT().Top().AnyTimes().Return(&b0)
	base.EXPECT().Length().AnyTimes().Return(uint64(1))

	Convey("Test of Block Cache", t, func() {
		Convey("Add:", func() {
			bc := NewBlockCache()
			bc.Add(&b0)
			b1node,_:=bc.Add(&b1)
			bc.Add(&b2)
			_,err := bc.Add(&b2)
			So(err, ShouldEqual, ErrDup)
			_,err = bc.Add(&b4)
			So(err, ShouldEqual, ErrNotFound)
			_,err = bc.Add(&b4)
			So(err, ShouldEqual, ErrDup)
			bc.Del(b1node)
		})

		Convey("Longest chain", func() {
			bc := NewBlockCache()

			bc.Add(&b1)
			b2node,_:=bc.Add(&b2)
			bc.Add(&b2a)
			So(bc.Head, ShouldEqual,b2node)
			b4node,_:=bc.Add(&b4)
			So(bc.Head, ShouldEqual,b4node)
		})

	})
}

