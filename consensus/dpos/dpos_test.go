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
)

func TestNewDPoS(t *testing.T) {
	Convey("Test fo NewDPos", t, func() {
		mockCtr := NewController(t)
		mockRouter := protocol_mock.NewMockRouter(mockCtr)

		//获取router实例
		guard := Patch(network.RouterFactory, func(_ string) (network.Router, error) {
			return mockRouter, nil
		})

		defer guard.Unpatch()

		heightChain := make(chan message.Message, 1)
		blkChain := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(heightChain, nil)

		mockRouter.EXPECT().FilteredChan(Any()).Return(blkChain, nil)

		txChan := make(chan message.Message, 1)
		//设置第一个通道txchan

		//构造测试数据
		txChan <- message.Message{
			Time:    20180426111111,
			From:    "0xaaaaaaaaaaaaa",
			To:      "0xbbbbbbbbbbbb",
			ReqType: 1,
			Body:    []byte{'a', 'b'}}

		//
		mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

		//设置第二个通道Blockchan
		blockChan := make(chan message.Message, 1)
		//设置第一个通道txchan
		//构造测试数据

		blockChan <- message.Message{
			Time:    20180426111111,
			From:    "0xaaaaaaaaaaaaa",
			To:      "0xbbbbbbbbbbbb",
			ReqType: 2,
			Body:    []byte{'c', 'd'}}
		//mockRouter.EXPECT().FilteredChan(network.Filter{
		//	WhiteList:  []message.Message{},
		//	BlackList:  []message.Message{},
		//	RejectType: []network.ReqType{},
		//	AcceptType: []network.ReqType{network.ReqNewBlock}}).Return(blockChan, nil)
		mockRouter.EXPECT().FilteredChan(Any()).Return(blockChan, nil)

		p, _ := NewDPoS(account.Account{"id0", []byte{23, 23, 23, 23, 23, 23}, []byte{23, 23}}, nil, nil, []string{})

		So(p.Account.GetId(), ShouldEqual, "id0")

	})

}

func TestDPoS_Run(t *testing.T) {
	Convey("Test fo Run", t, func() {
		mockCtr := NewController(t)
		mockRouter := protocol_mock.NewMockRouter(mockCtr)

		//获取router实例
		guard := Patch(network.RouterFactory, func(_ string) (network.Router, error) {
			return mockRouter, nil
		})

		defer guard.Unpatch()

		heightChan := make(chan message.Message, 1)
		blkChan := make(chan message.Message, 1)
		mockRouter.EXPECT().FilteredChan(Any()).Return(heightChan, nil)
		mockRouter.EXPECT().FilteredChan(Any()).Return(blkChan, nil)

		txChan := make(chan message.Message, 1)
		//设置第一个通道txchan

		main := lua.NewMethod("main", 0, 1)
		code := `function main()
						Put("hello", "world")
						return "success"
					end`
		lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

		b0 := block.Block{
			Head: block.BlockHead{
				Version:    0,
				ParentHash: []byte("nothing"),
				Witness:    "w0",
			},
			Content: []tx.Tx{tx.NewTx(0, &lc)},
		}

		tx:=tx.NewTx(0, &lc)

		//构造测试数据
		txChan <- message.Message{
			Time:    20180426111111,
			From:    "0xaaaaaaaaaaaaa",
			To:      "0xbbbbbbbbbbbb",
			ReqType: 1,
			Body:    tx.Encode()}

		mockRouter.EXPECT().FilteredChan(Any()).Return(txChan, nil)

		//设置第二个通道Blockchan
		blockChan := make(chan message.Message, 1)
		//设置第一个通道txchan
		//构造测试数据

		blockChan <- message.Message{
			Time:    20180426111111,
			From:    "0xaaaaaaaaaaaaa",
			To:      "0xbbbbbbbbbbbb",
			ReqType: 2,
			Body:    b0.Encode()}
		mockRouter.EXPECT().FilteredChan(Any()).Return(blockChan, nil)

		p, _ := NewDPoS(account.Account{"id0", []byte{23, 23, 23, 23, 23, 23}, []byte{23, 23}}, nil, nil, []string{"id1"})

		p.Run()
		p.Stop()

		//time.Sleep(20 * time.Second)
	})

}
