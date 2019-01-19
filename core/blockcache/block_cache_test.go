package blockcache

import (
	"encoding/json"
	"testing"

	. "github.com/golang/mock/gomock"
	core_mock "github.com/iost-official/go-iost/core/mocks"
	db_mock "github.com/iost-official/go-iost/db/mocks"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/vm/database"
	. "github.com/smartystreets/goconvey/convey"
)

func genBlock(fa *block.Block, wit string, num uint64) *block.Block {
	ret := &block.Block{
		Head: &block.BlockHead{
			Witness: wit,
			Number:  int64(num),
		},
	}
	if fa == nil {
		ret.Head.ParentHash = []byte("Im a single block")
	} else {
		ret.Head.ParentHash = fa.HeadHash()
	}
	ret.CalculateHeadHash()
	return ret
}
func TestBlockCache(t *testing.T) {
	ctl := NewController(t)
	b0 := &block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: []byte("nothing"),
			Witness:    "w0",
			Number:     0,
		},
	}
	b0.CalculateHeadHash()
	b1 := genBlock(b0, "w1", 1)
	b2 := genBlock(b1, "w2", 2)
	b2a := genBlock(b1, "w3", 3)
	b3 := genBlock(b2, "w4", 4)
	b4 := genBlock(b2a, "w5", 5)
	b3a := genBlock(b2, "w6", 6)
	b5 := genBlock(b3a, "w7", 7)

	s1 := genBlock(nil, "w1", 1)
	s2 := genBlock(s1, "w2", 2)
	s2a := genBlock(s1, "w3", 3)
	s3 := genBlock(s2, "w4", 4)
	statedb := db_mock.NewMockMVCCDB(ctl)
	statedb.EXPECT().Flush(Any()).AnyTimes().Return(nil)
	statedb.EXPECT().Fork().AnyTimes().Return(statedb)
	statedb.EXPECT().Checkout(Any()).AnyTimes().Return(true)
	statedb.EXPECT().Size().AnyTimes().Return(int64(10000), nil)

	statedb.EXPECT().Get("state", "b-vote_producer.iost-"+"pendingBlockNumber").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal("1"), nil
	})
	statedb.EXPECT().Get("state", "b-vote_producer.iost-"+"pendingProducerList").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal("[\"aaaa\",\"bbbbb\"]"), nil
	})
	statedb.EXPECT().Get("snapshot", "blockHead").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		bhJson, _ := json.Marshal(b0.Head)
		return string(bhJson), nil
	})
	//"m-vote_producer.iost-producerTable"
	statedb.EXPECT().Get("state", Any()).AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(`{"loc":"11","url":"22","netId":"33","online":true,"score":0,"votes":0}`), nil
	})

	base := core_mock.NewMockChain(ctl)
	base.EXPECT().Top().AnyTimes().Return(b0, nil)
	base.EXPECT().Push(Any()).AnyTimes().Return(nil)
	base.EXPECT().TxTotal().AnyTimes().Return(int64(10))
	base.EXPECT().Size().AnyTimes().Return(int64(10000), nil)
	global := core_mock.NewMockBaseVariable(ctl)
	global.EXPECT().BlockChain().AnyTimes().Return(base)
	global.EXPECT().StateDB().AnyTimes().Return(statedb)
	config := common.Config{
		DB: &common.DBConfig{
			LdbPath: "./",
		},
		Snapshot: &common.SnapshotConfig{
			Enable: false,
		},
	}
	global.EXPECT().Config().AnyTimes().Return(&config)
	Convey("Test of Block Cache", t, func() {
		Convey("Add:", func() {
			CleanBlockCacheWAL()
			bc, _ := NewBlockCache(global)
			defer bc.CleanDir()
			//fmt.Printf("Leaf:%+v\n",bc.Leaf)
			_ = bc.Add(b1)
			//fmt.Printf("Leaf:%+v\n",bc.Leaf)
			//bc.Draw()
			bc.Add(b2)
		})

		Convey("Flush", func() {
			CleanBlockCacheWAL()
			bc, _ := NewBlockCache(global)
			defer bc.CleanDir()
			bc.Add(b1)
			//bc.Draw()
			bc.Add(b2)
			//bc.Draw()
			bc.Add(b2a)
			//bc.Draw()
			bc.Add(b3)
			//bc.Draw()
			//b4node, _ := bc.Add(b4)
			//bc.Draw()
			bc.Add(b3a)
			//bc.Draw()
			bc.Add(b5)
			//bc.Draw()

			bc.Add(s1)
			bc.Add(s2)
			bc.Add(s2a)
			bc.Add(s3)
			//bc.Draw()
			//bc.Flush(b4node)
			//bc.Draw()

		})

		Convey("GetBlockbyNumber", func() {
			CleanBlockCacheWAL()
			bc, _ := NewBlockCache(global)
			defer bc.CleanDir()
			b1node := bc.Add(b1)
			bc.Link(b1node)
			//bc.Draw()
			b2node := bc.Add(b2)
			bc.Link(b2node)
			// bc.Draw()
			b2anode := bc.Add(b2a)
			bc.Link(b2anode)
			// bc.Draw()
			b3node := bc.Add(b3)
			bc.Link(b3node)
			// bc.Draw()
			b4node := bc.Add(b4)
			bc.Link(b4node)
			// bc.Draw()
			b3anode := bc.Add(b3a)
			bc.Link(b3anode)
			// bc.Draw()
			b5node := bc.Add(b5)
			bc.Link(b5node)
			// bc.Draw()
			So(bc.head, ShouldEqual, b5node)
			blk, _ := bc.GetBlockByNumber(7)
			So(blk, ShouldEqual, b5node.Block)
			blk, _ = bc.GetBlockByNumber(6)
			So(blk, ShouldEqual, b3anode.Block)
			blk, _ = bc.GetBlockByNumber(2)
			So(blk, ShouldEqual, b2node.Block)
			blk, _ = bc.GetBlockByNumber(1)
			So(blk, ShouldEqual, b1node.Block)
			blk, _ = bc.GetBlockByNumber(4)
			So(blk, ShouldEqual, nil)

			bc.Flush(b4node)
			//bc.Draw()

		})

	})
}

func TestVote(t *testing.T) {
	ctl := NewController(t)
	b0 := &block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: []byte("nothing"),
			Witness:    "w0",
			Number:     0,
		},
	}
	b0.CalculateHeadHash()

	b1 := genBlock(b0, "w1", 1)
	b2 := genBlock(b1, "w2", 2)
	b3 := genBlock(b2, "w3", 3)
	//b4 := genBlock(b3, "w4", 4)
	//b5 := genBlock(b4, "w5", 5)
	//
	//fmt.Println(b5)

	statedb := db_mock.NewMockMVCCDB(ctl)
	statedb.EXPECT().Flush(Any()).AnyTimes().Return(nil)
	statedb.EXPECT().Fork().AnyTimes().Return(statedb)
	statedb.EXPECT().Checkout(Any()).AnyTimes().Return(true)

	tpl := "[\"a1\",\"a2\",\"a3\",\"a4\",\"a5\"]"
	//tpl1 := "[\"b1\",\"b2\",\"b3\",\"b4\",\"b5\"]"
	statedb.EXPECT().Get("state", "b-vote_producer.iost-"+"pendingBlockNumber").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal("5"), nil
	})
	statedb.EXPECT().Get("state", "b-vote_producer.iost-"+"pendingProducerList").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(tpl), nil
	})
	statedb.EXPECT().Get("snapshot", "blockHead").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		bhJson, _ := json.Marshal(b0.Head)
		return string(bhJson), nil
	})
	statedb.EXPECT().Get("state", Any()).AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(`{"loc":"11","url":"22","netId":"33","online":true,"score":0,"votes":0}`), nil
	})

	base := core_mock.NewMockChain(ctl)
	base.EXPECT().Top().AnyTimes().Return(b0, nil)
	base.EXPECT().Push(Any()).AnyTimes().Return(nil)
	global := core_mock.NewMockBaseVariable(ctl)
	global.EXPECT().BlockChain().AnyTimes().Return(base)
	global.EXPECT().StateDB().AnyTimes().Return(statedb)
	config := common.Config{
		DB: &common.DBConfig{
			LdbPath: "./",
		},
		Snapshot: &common.SnapshotConfig{
			Enable: false,
		},
	}
	global.EXPECT().Config().AnyTimes().Return(&config)

	Convey("test api", t, func() {
		var wl WitnessList
		pl := []string{"p1", "p2", "p3"}
		var pn int64 = 1
		al := []string{"a1", "a2", "a3"}

		wl.SetPending(pl)
		So(StringSliceEqual(pl, wl.Pending()), ShouldBeTrue)
		wl.SetPendingNum(pn)
		So(wl.PendingNum(), ShouldEqual, pn)
		wl.SetActive(al)
		So(StringSliceEqual(al, wl.Active()), ShouldBeTrue)

	})
	Convey("test update", t, func() {
		bc, _ := NewBlockCache(global)
		defer bc.CleanDir()
		//fmt.Printf("Leaf:%+v\n",bc.Leaf)
		node1 := NewBCN(bc.linkedRoot, b1)
		node2 := NewBCN(node1, b2)
		node3 := NewBCN(node2, b3)
		bc.Link(node1)
		So(StringSliceEqual([]string{"a1", "a2", "a3", "a4", "a5"}, bc.head.Pending()), ShouldBeTrue)
		bc.Link(node2)
		So(StringSliceEqual([]string{"a1", "a2", "a3", "a4", "a5"}, bc.head.Pending()), ShouldBeTrue)
		bc.Link(node3)
		So(StringSliceEqual([]string{"a1", "a2", "a3", "a4", "a5"}, bc.head.Pending()), ShouldBeTrue)

	})
	Convey("test info", t, func() {
		bc, _ := NewBlockCache(global)
		defer bc.CleanDir()
		for _, v := range bc.linkedRoot.NetID() {
			So("33", ShouldEqual, v)
		}
	})
}

func StringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
