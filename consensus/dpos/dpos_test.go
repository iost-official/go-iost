package dpos

import (
	"testing"

	. "github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/network/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/mocks"
	"time"
)

func TestNewDPoS(t *testing.T) {
	Convey("Test fo NewDPos", t, func() {
		mockCtr := NewController(t)
		mockRouter := protocol_mock.NewMockRouter(mockCtr)
		mockBc := core_mock.NewMockChain(mockCtr)
		mockPool := core_mock.NewMockPool(mockCtr)

		//获取router实例
		guard := Patch(network.RouterFactory, func(_ string) (network.Router, error) {
			return mockRouter, nil
		})

		defer guard.Unpatch()

		heightChan := make(chan message.Message, 1)
		blkChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(heightChan, nil)
		mockRouter.EXPECT().FilteredChan(Any()).Return(blkChan, nil)

		// 设置第一个通道txchan
		txChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

		// 设置第二个通道Blockchan
		blockChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(blockChan, nil)

		// 创世块的询问和插入
		var getNumber uint64
		var pushNumber int64
		mockBc.EXPECT().GetBlockByNumber(Any()).Do(func(number uint64) *block.Block {
			getNumber = number
			return nil
		})
		mockBc.EXPECT().Push(Any()).Do(func(block *block.Block) error {
			pushNumber = block.Head.Number
			return nil
		})

		p, _ := NewDPoS(account.Account{"id0", []byte("pubkey"), []byte("seckey")}, mockBc, mockPool, []string{})

		So(p.Account.GetId(), ShouldEqual, "id0")
		So(getNumber, ShouldEqual, 0)
		So(pushNumber, ShouldEqual,0)
	})

}

func TestRunGenerateBlock(t *testing.T) {
	Convey("Test of Run (Generate Block)", t, func() {
		mockCtr := NewController(t)
		mockRouter := protocol_mock.NewMockRouter(mockCtr)
		mockBc := core_mock.NewMockChain(mockCtr)
		mockPool := core_mock.NewMockPool(mockCtr)

		//获取router实例
		guard := Patch(network.RouterFactory, func(_ string) (network.Router, error) {
			return mockRouter, nil
		})

		defer guard.Unpatch()

		heightChan := make(chan message.Message, 1)
		blkChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(heightChan, nil)
		mockRouter.EXPECT().FilteredChan(Any()).Return(blkChan, nil)

		//设置第一个通道txchan
		txChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

		//设置第二个通道Blockchan
		blockChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(blockChan, nil)

		mockBc.EXPECT().GetBlockByNumber(Eq(uint64(0))).Return(nil)
		var genesis *block.Block
		mockBc.EXPECT().Push(Any()).Do(func(block *block.Block) error {
			genesis = block
			return nil
		})
		p, _ := NewDPoS(account.Account{"id0", []byte("pubkey"), []byte("seckey")}, mockBc, mockPool, []string{"id0", "id1", "id2"})

		main := lua.NewMethod("main", 0, 1)
		code := `function main()
						Put("hello", "world")
						return "success"
					end`
		lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)
		newTx:=tx.NewTx(0, &lc)
		//构造测试数据
		txChan <- message.Message{
			Time:    20180426111111,
			From:    "0xaaaaaaaaaaaaa",
			To:      "0xbbbbbbbbbbbb",
			ReqType: 1,
			Body:    newTx.Encode()}

		mockBc.EXPECT().Top().Return(genesis).AnyTimes()

		var blk block.Block
		var reqType network.ReqType
		mockRouter.EXPECT().Broadcast(Any()).Do(func(req message.Message) {
			reqType = network.ReqType(req.ReqType)
			blk.Decode(req.Body)
		}).AnyTimes()
		p.Run()

		time.Sleep(time.Second * 2)
		So(reqType, ShouldEqual, network.ReqNewBlock)
		So(blk.Head.Number, ShouldEqual, 1)
		So(string(blk.Head.ParentHash), ShouldEqual, string(genesis.Head.Hash()))
		So(blk.Head.Witness, ShouldEqual, "id0")

		p.Stop()

	})

}
