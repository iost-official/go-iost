package dpos

import (
	"testing"

	. "github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/core"
	"github.com/iost-official/prototype/p2p"
	"github.com/iost-official/prototype/p2p/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"time"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/common"
)

func TestNewDPoS(t *testing.T) {
	Convey("Test fo NewDPos", t, func() {
		mockCtr := NewController(t)
		mockRouter := mock_p2p.NewMockRouter(mockCtr)

		//获取router实例
		guard := Patch(p2p.RouterFactory, func(_ string) (p2p.Router, error) {
			return mockRouter, nil
		})

		defer guard.Unpatch()

		txChan := make(chan core.Request, 1)
		//设置第一个通道txchan
		type Request struct {
			Time    int64
			From    string
			To      string
			ReqType int
			Body    []byte
		}
		//构造测试数据
		txChan <- core.Request{
			Time:    20180426111111,
			From:    "0xaaaaaaaaaaaaa",
			To:      "0xbbbbbbbbbbbb",
			ReqType: 1,
			Body:    []byte{'a', 'b'}}

		mockRouter.EXPECT().FilteredChan(p2p.Filter{
			WhiteList:  []core.Member{},
			BlackList:  []core.Member{},
			RejectType: []p2p.ReqType{},
			AcceptType: []p2p.ReqType{
				p2p.ReqPublishTx,
				ReqTypeVoteTest, // Only for test
			}}).Return(txChan,nil)

		//设置第二个通道Blockchan
		blockChan := make(chan core.Request, 1)
		//设置第一个通道txchan
		//构造测试数据

		blockChan<-core.Request{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:2,
			Body:[]byte{'c','d'}}
		mockRouter.EXPECT().FilteredChan(p2p.Filter{
			WhiteList:  []core.Member{},
			BlackList:  []core.Member{},
			RejectType: []p2p.ReqType{},
			AcceptType: []p2p.ReqType{
				p2p.ReqNewBlock}}).Return(blockChan,nil)

		p, _ := NewDPoS(core.Member{"id0", []byte{23, 23, 23, 23, 23, 23}, []byte{23, 23}}, nil)

		p.Genesis(core.Timestamp{}, []byte{})
	})

}

func TestDPoS_Run(t *testing.T) {
	Convey("Test fo Run", t, func() {
		mockCtr := NewController(t)
		mockRouter := mock_p2p.NewMockRouter(mockCtr)

		//获取router实例
		guard := Patch(p2p.RouterFactory, func(_ string) (p2p.Router, error) {
			return mockRouter, nil
		})

		defer guard.Unpatch()

		txChan := make(chan core.Request, 1)
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
		lac := new(vm.LuaContract)
		//初始化交易数据
		var txData core.Tx
		txData = core.Tx{
			Time:     time.Now().Unix(),
			Contract: lac,
			Signs:[]common.Signature{},
			Publisher:[]common.Signature{}}

		txChan<-core.Request{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:int(p2p.ReqPublishTx),
			Body:txData.Encode()}

		mockRouter.EXPECT().FilteredChan(p2p.Filter{
			WhiteList:  []core.Member{},
			BlackList:  []core.Member{},
			RejectType: []p2p.ReqType{},
			AcceptType: []p2p.ReqType{
				p2p.ReqPublishTx,
				ReqTypeVoteTest, // Only for test
			}}).Return(txChan,nil)

		//设置第二个通道Blockchan
		blockChan :=make(chan core.Request,1)
		var blockData core.Block
		blockData = core.Block{
			Version:1.0,
			Head:core.BlockHead{
				Version:1,
				ParentHash:[]byte{'a','b'},
				TreeHash:[]byte{'a','b'},
				BlockHash:[]byte{'a','b'},
				Info:[]byte{'a','b'},
				Number:33,
				Witness:"test",
				Signature:[]byte{'a','b'},
				Time:core.Timestamp{Slot:1111}}}

		//构造测试数据
		blockChan<-core.Request{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:int(p2p.ReqNewBlock),
			Body:blockData.Encode()}

		mockRouter.EXPECT().FilteredChan(p2p.Filter{
			WhiteList:  []core.Member{},
			BlackList:  []core.Member{},
			RejectType: []p2p.ReqType{},
			AcceptType: []p2p.ReqType{
				p2p.ReqNewBlock}}).Return(blockChan,nil)

		p, _ := NewDPoS(core.Member{"id0", []byte{23, 23, 23, 23, 23, 23}, []byte{23, 23}}, nil)

		p.Run()

		time.Sleep(20 * time.Second)
	})

}
