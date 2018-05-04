package dpos

import (
	"testing"

	. "github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"

	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/network/mocks"
	"github.com/iost-official/prototype/vm"
	. "github.com/smartystreets/goconvey/convey"
	"time"
	"github.com/iost-official/prototype/common"
	. "github.com/iost-official/prototype/consensus/common"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/tx"
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

		txChan := make(chan message.Message, 1)
		//设置第一个通道txchan

		//构造测试数据
		txChan <- message.Message{
			Time:    20180426111111,
			From:    "0xaaaaaaaaaaaaa",
			To:      "0xbbbbbbbbbbbb",
			ReqType: 1,
			Body:    []byte{'a', 'b'}}

		mockRouter.EXPECT().FilteredChan(network.Filter{
			WhiteList:  []message.Message{},
			BlackList:  []message.Message{},
			RejectType: []network.ReqType{},
			AcceptType: []network.ReqType{
				network.ReqPublishTx,
				ReqTypeVoteTest, // Only for test
			}}).Return(txChan, nil)

		//设置第二个通道Blockchan
		blockChan := make(chan message.Message, 1)
		//设置第一个通道txchan
		//构造测试数据

		blockChan<-message.Message{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:2,
			Body:[]byte{'c','d'}}
		mockRouter.EXPECT().FilteredChan(network.Filter{
			WhiteList:  []message.Message{},
			BlackList:  []message.Message{},
			RejectType: []network.ReqType{},
			AcceptType: []network.ReqType{
				network.ReqNewBlock}}).Return(blockChan,nil)


		p, _ := NewDPoS(account.Account{"id0", []byte{23, 23, 23, 23, 23, 23}, []byte{23, 23}}, nil)

		p.Genesis(Timestamp{}, []byte{})
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

		txChan := make(chan message.Message, 1)
		//设置第一个通道txchan
		type Request struct {
			Time    int64
			From    string
			To      string
			ReqType int
			Body    []byte
		}

		//构造交易测试数据
		//构造智能合约
		lac := new(vm.Contract)
		//初始化交易数据

		var txData tx.Tx
		txData = tx.Tx{
			Time:     time.Now().Unix(),
			Contract: lac,
			Signs:[]common.Signature{},
			Publisher:common.Signature{}}

		txChan<-message.Message{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:int32(network.ReqPublishTx),
			Body:txData.Encode()}


		mockRouter.EXPECT().FilteredChan(network.Filter{
			WhiteList:  []message.Message{},
			BlackList:  []message.Message{},
			RejectType: []network.ReqType{},
			AcceptType: []network.ReqType{
				network.ReqPublishTx,
				ReqTypeVoteTest, // Only for test
			}}).Return(txChan, nil)

		//设置第二个通道Blockchan

		blockChan :=make(chan message.Message,1)
		var blockData block.Block
		blockData = block.Block{
			Version:1.0,
			Head:block.BlockHead{
				Version:1,
				ParentHash:[]byte{'a','b'},
				TreeHash:[]byte{'a','b'},
				BlockHash:[]byte{'a','b'},
				Info:[]byte{'a','b'},
				Number:33,
				Witness:"test",
				Signature:[]byte{'a','b'},
				Time: 1111}}

		//构造测试数据
		blockChan<-message.Message{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:int32(network.ReqNewBlock),
			Body:blockData.Encode()}

		mockRouter.EXPECT().FilteredChan(network.Filter{
			WhiteList:  []message.Message{},
			BlackList:  []message.Message{},
			RejectType: []network.ReqType{},
			AcceptType: []network.ReqType{
				network.ReqNewBlock}}).Return(blockChan,nil)


		p, _ := NewDPoS(account.Account{"id0", []byte{23, 23, 23, 23, 23, 23}, []byte{23, 23}}, nil)

		p.Run()

		time.Sleep(20 * time.Second)
	})

}
