package consensus_common

import (
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/iost-official/prototype/core/tx"
)

func TestBlockCachePoW(t *testing.T) {
	b0 := block.Block{
		Head: block.BlockHead{
			Version: 1,
			ParentHash: []byte("nothing"),
		},
		Content: []tx.Tx{tx.NewTx(0, nil)},
	}

	b1 := block.Block{
		Head: block.BlockHead{
			Version: 1,
			ParentHash: b0.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(1, nil)},
	}

	b2 := block.Block{
		Head: block.BlockHead{
			Version: 1,
			ParentHash: b1.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(2, nil)},
	}

	b2a := block.Block{
		Head: block.BlockHead{
			Version: 1,
			ParentHash: b1.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(-2, nil)},
	}

	b3 := block.Block{
		Head: block.BlockHead{
			Version: 1,
			ParentHash: b2.HeadHash(),
		},
		Content: []tx.Tx{tx.NewTx(3, nil)},
	}

	b4 := block.Block{
		Head: block.BlockHead{
			Version: 1,
			ParentHash: b3.HeadHash(),
		},
	}

	ctl := gomock.NewController(t)

	verifier := func(blk *block.Block, chain block.Chain) (bool, state.Pool) {
		return true, nil
	}

	base := core_mock.NewMockChain(ctl)
	base.EXPECT().Top().AnyTimes().Return(&b0)

	Convey("Test of Block Cache (PoW)", t, func() {
		Convey("Add:", func() {
			Convey("normal:", func() {
				bc := NewBlockCache(base, 4)
				err := bc.Add(&b1, verifier)
				So(err, ShouldBeNil)
				So(bc.cachedRoot.depth, ShouldEqual, 1)

			})

			Convey("fork and error", func() {
				bc := NewBlockCache(base, 4)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				bc.Add(&b2a, verifier)
				So(bc.cachedRoot.depth, ShouldEqual, 2)

				verifier = func(blk *block.Block, chain block.Chain) (bool, state.Pool) {
					return false, nil
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
				verifier = func(blk *block.Block, chain block.Chain) (bool, state.Pool) {
					return true, nil
				}
				bc := NewBlockCache(base, 3)
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
				bc := NewBlockCache(base, 10)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				ans := bc.LongestChain().Top().Content[0].Nonce
				So(ans, ShouldEqual, 2)
			})

			Convey("forked", func() {
				var bc BlockCache = NewBlockCache(base, 10)

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
			bc := NewBlockCache(base, 10)
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

func TestBlockCacheDPoS(t *testing.T) {
	b0 := block.Block{
		Head: block.BlockHead{
			Version: 0,
			ParentHash: []byte("nothing"),
			Witness: "w0",
		},
		Content: []tx.Tx{tx.NewTx(0, nil)},
	}

	b1 := block.Block{
		Head: block.BlockHead{
			Version: 0,
			ParentHash: b0.HeadHash(),
			Witness: "w1",
		},
		Content: []tx.Tx{tx.NewTx(1, nil)},
	}

	b2 := block.Block{
		Head: block.BlockHead{
			Version: 0,
			ParentHash: b1.HeadHash(),
			Witness: "w2",
		},
		Content: []tx.Tx{tx.NewTx(2, nil)},
	}

	b2a := block.Block{
		Head: block.BlockHead{
			Version: 0,
			ParentHash: b1.HeadHash(),
			Witness: "w3",
		},
		Content: []tx.Tx{tx.NewTx(-2, nil)},
	}

	b3 := block.Block{
		Head: block.BlockHead{
			Version: 0,
			ParentHash: b2.HeadHash(),
			Witness: "w1",
		},
		Content: []tx.Tx{tx.NewTx(3, nil)},
	}

	b4 := block.Block{
		Head: block.BlockHead{
			Version: 0,
			ParentHash: b2a.HeadHash(),
			Witness: "w2",
		},
		Content: []tx.Tx{tx.NewTx(4, nil)},
	}

	ctl := gomock.NewController(t)

	verifier := func(blk *block.Block, chain block.Chain) (bool, state.Pool) {
		return true, nil
	}

	base := core_mock.NewMockChain(ctl)
	base.EXPECT().Top().AnyTimes().Return(&b0)

	Convey("Test of Block Cache (DPoS)", t, func() {
		Convey("Add:", func() {
			Convey("auto push", func() {
				var ans int64
				base.EXPECT().Push(gomock.Any()).AnyTimes().Do(func(block *block.Block) error {
					ans = block.Content[0].Nonce
					return nil
				})
				bc := NewBlockCache(base, 2)
				bc.Add(&b1, verifier)
				bc.Add(&b2, verifier)
				bc.Add(&b2a, verifier)
				bc.Add(&b3, verifier)
				bc.Add(&b4, verifier)
				So(ans, ShouldEqual, 1)
			})
		})

	})

}
