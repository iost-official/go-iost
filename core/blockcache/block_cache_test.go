package blockcache

import (
	"testing"

	"errors"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/db/mocks"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"

	"github.com/iost-official/prototype/core/block"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBlockCachePoW(t *testing.T) {
	ctl := gomock.NewController(t)
	pool := core_mock.NewMockPool(ctl)

	pool.EXPECT().Flush().AnyTimes().Return(nil)

	main := lua.NewMethod(vm.Public, "main", 0, 1)
	code := `function main()
						Put("hello", "world")
						return "success"
					end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

	b0 := block.Block{
		Head: block.BlockHead{
			Version:    1,
			ParentHash: []byte("nothing"),
		},
		Content: []tx.Tx{tx.NewTx(0, &lc)},
	}

	b1 := block.Block{
		Head: block.BlockHead{
			Version:    1,
			ParentHash: b0.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(1, &lc)},
	}

	b2 := block.Block{
		Head: block.BlockHead{
			Version:    1,
			ParentHash: b1.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(2, &lc)},
	}

	b2a := block.Block{
		Head: block.BlockHead{
			Version:    1,
			ParentHash: b1.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(-2, &lc)},
	}

	b3 := block.Block{
		Head: block.BlockHead{
			Version:    1,
			ParentHash: b2.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(3, &lc)},
	}

	b4 := block.Block{
		Head: block.BlockHead{
			Version:    1,
			ParentHash: b3.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(4, &lc)},
	}

	verifier := func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
		return pool, nil
	}

	base := core_mock.NewMockChain(ctl)
	base.EXPECT().Top().AnyTimes().Return(&b0)
	base.EXPECT().Length().AnyTimes().Return(uint64(1))

	Convey("Test of Block Cache (PoW)", t, func() {
		Convey("Add:", func() {
			Convey("normal:", func() {
				bc := NewBlockCache(base, pool, 4)
				err := bc.Add(&b1, verifier)
				So(err, ShouldBeNil)
				So(bc.cachedRoot.bc.depth, ShouldEqual, 1)

			})

			Convey("fork and error", func() {
				bc := NewBlockCache(base, pool, 4)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				bc.Add(&b2a, verifier)
				So(bc.cachedRoot.bc.depth, ShouldEqual, 2)

				verifier = func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
					return nil, errors.New("test")
				}
				err := bc.Add(&b3, verifier)
				So(err, ShouldNotBeNil)
			})

			Convey("auto push", func() {
				var ans int64
				base.EXPECT().Push(gomock.Any()).AnyTimes().Do(func(block *block.Block) error {
					ans = block.Content[0].Nonce
					return nil
				})
				verifier = func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
					return pool, nil
				}
				bc := NewBlockCache(base, pool, 3)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				bc.Add(&b2a, verifier)
				bc.Add(&b3, verifier)
				bc.Add(&b4, verifier)
				So(ans, ShouldEqual, 1)
			})
		})

		Convey("Longest chain", func() {
			Convey("no forked", func() {
				bc := NewBlockCache(base, pool, 10)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				ans := bc.LongestChain().Top().Content[0].Nonce
				So(ans, ShouldEqual, 2)
			})

			Convey("forked", func() {
				var bc BlockCache = NewBlockCache(base, pool, 10)

				bc.Add(&b1, verifier)
				bc.Add(&b2a, verifier)
				bc.Add(&b2, verifier)
				ans := bc.LongestChain().Top().Content[0].Nonce
				So(ans, ShouldEqual, -2)
				bc.Add(&b3, verifier)
				ans = bc.LongestChain().Top().Content[0].Nonce
				So(ans, ShouldEqual, 3)
			})
		})

		Convey("find blk", func() {
			bc := NewBlockCache(base, pool, 10)
			bc.Add(&b1, verifier)
			bc.Add(&b2a, verifier)
			bc.Add(&b2, verifier)
			ans, err := bc.FindBlockInCache(b2a.HeadHash())
			So(err, ShouldBeNil)
			So(ans.Content[0].Nonce, ShouldEqual, -2)

			ans, err = bc.FindBlockInCache(b3.HeadHash())
			So(err, ShouldNotBeNil)

		})

	})

}

func TestBlockCachePoB(t *testing.T) {
	ctl := gomock.NewController(t)
	pool := core_mock.NewMockPool(ctl)

	pool.EXPECT().Flush().AnyTimes().Return(nil)

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

	verifier := func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
		return pool, nil
	}

	base := core_mock.NewMockChain(ctl)
	base.EXPECT().Top().AnyTimes().Return(&b0)
	base.EXPECT().Length().AnyTimes().Return(uint64(1))

	Convey("Test of Block Cache (PoB)", t, func() {
		Convey("Add:", func() {
			var ans int64
			base.EXPECT().Push(gomock.Any()).Do(func(block *block.Block) error {
				ans = block.Content[0].Nonce
				return nil
			})
			Convey("auto push", func() {
				ans = 0
				bc := NewBlockCache(base, pool, 2)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				bc.Add(&b2a, verifier)
				bc.Add(&b3, verifier)
				bc.Add(&b4, verifier)
				So(ans, ShouldEqual, 1)
			})

			Convey("deal with singles", func() {
				ans = 0
				bc := NewBlockCache(base, pool, 2)
				bc.Add(&b2, verifier)
				bc.Add(&b2a, verifier)
				bc.Add(&b3, verifier)
				bc.Add(&b4, verifier)
				So(len(bc.singleBlockRoot.children), ShouldEqual, 2)
				bc.Add(&b1, verifier)
				//bc.AddSingles(verifier)
				So(len(bc.singleBlockRoot.children), ShouldEqual, 0)
				So(ans, ShouldEqual, 1)
			})

			Convey("deal with dups", func() {
				bc := NewBlockCache(base, pool, 2)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				err := bc.Add(&b2, verifier)
				So(err, ShouldEqual, ErrDup)
				err = bc.Add(&b4, verifier)
				So(err, ShouldEqual, ErrNotFound)
				err = bc.Add(&b4, verifier)
				So(err, ShouldEqual, ErrDup)
			})
		})

		Convey("Longest chain", func() {
			Convey("no forked", func() {
				bc := NewBlockCache(base, pool, 10)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				ans := bc.LongestChain().Top().Content[0].Nonce
				So(ans, ShouldEqual, 2)
			})

			Convey("forked", func() {
				var bc BlockCache = NewBlockCache(base, pool, 10)

				bc.Add(&b1, verifier)
				bc.Add(&b2a, verifier)
				bc.Add(&b2, verifier)
				ans := bc.LongestChain().Top().Content[0].Nonce
				So(ans, ShouldEqual, -2)
				bc.Add(&b3, verifier)
				ans = bc.LongestChain().Top().Content[0].Nonce
				So(ans, ShouldEqual, 3)
			})
		})

	})
}

func TestStatePool(t *testing.T) {
	Convey("Test of verifier", t, func() {
		ctl := gomock.NewController(t)
		mockDB := db_mock.NewMockDatabase(ctl)
		mockDB.EXPECT().GetHM(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.New("not found"))
		mockDB.EXPECT().Get(gomock.Any()).AnyTimes().Return(nil, errors.New("not found"))
		db := state.NewDatabase(mockDB)
		pool := state.NewPool(db)
		pool.Put(state.Key("a"), state.MakeVInt(int(0)))

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
				ParentHash: b3.HeadHash(),
				Witness:    "w2",
			},
			Content: []tx.Tx{tx.NewTx(4, &lc)},
		}

		verifier := func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
			p := pool.Copy()
			p.Put(state.Key("a"), state.MakeVInt(int(blk.Content[0].Nonce)))
			return p, nil
		}

		base := core_mock.NewMockChain(ctl)
		base.EXPECT().Top().AnyTimes().Return(&b0)
		var ans int64
		base.EXPECT().Push(gomock.Any()).Do(func(block *block.Block) error {
			ans = block.Content[0].Nonce
			return nil
		})

		bc := NewBlockCache(base, pool, 10)

		bc.Add(&b1, verifier)
		bc.Add(&b2, verifier)
		bc.Add(&b2a, verifier)
		bc.Add(&b3, verifier)
		bc.Add(&b4, verifier)

		lp := bc.LongestPool()
		v, err := lp.Get(state.Key("a"))
		So(err, ShouldBeNil)
		So(v.(*state.VInt).ToInt(), ShouldEqual, 4)
		bp := bc.BasePool()
		v, err = bp.Get(state.Key("a"))
		So(err, ShouldBeNil)
		So(v.(*state.VInt).ToInt(), ShouldEqual, 0)

	})
}
