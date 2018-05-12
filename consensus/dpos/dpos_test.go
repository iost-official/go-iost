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
				reqTypeVoteTest, // Only for test
			}}).Return(txChan, nil)

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
		mockRouter.EXPECT().FilteredChan(network.Filter{
			WhiteList:  []message.Message{},
			BlackList:  []message.Message{},
			RejectType: []network.ReqType{},
			AcceptType: []network.ReqType{
				network.ReqNewBlock}}).Return(blockChan, nil)

		p, _ := NewDPoS(account.Account{"id0", []byte{23, 23, 23, 23, 23, 23},[]byte{23, 23}}, nil,nil, []string{})

		p.genesis(0)
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
				reqTypeVoteTest, // Only for test
			}}).Return(txChan, nil)

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
		mockRouter.EXPECT().FilteredChan(network.Filter{
			WhiteList:  []message.Message{},
			BlackList:  []message.Message{},
			RejectType: []network.ReqType{},
			AcceptType: []network.ReqType{
				network.ReqNewBlock}}).Return(blockChan, nil)

		p, _ := NewDPoS(account.Account{"id0", []byte{23, 23, 23, 23, 23, 23}, []byte{23, 23}}, nil,nil, []string{})

		p.Run()

		//time.Sleep(20 * time.Second)
	})

}
