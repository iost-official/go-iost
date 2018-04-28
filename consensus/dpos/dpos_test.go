package dpos

import (
	"testing"

	"github.com/iost-official/prototype/core"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/p2p/mocks"
	. "github.com/bouk/monkey"
	"github.com/iost-official/prototype/p2p"
	"time"
)

func TestDPoS(t *testing.T) {
	Convey("Test of DPos", t, func() {
		dpos, _ := NewDPoS(core.Member{"id0", []byte{}, []byte{}}, nil)
		dpos.Run()
	})

}

func TestNewDPoS(t *testing.T) {
	Convey("Test fo NewDPos", t, func() {
		mockCtr :=NewController(t)
		mockRouter:=mock_p2p.NewMockRouter(mockCtr)

		//获取router实例
		guard:= Patch(p2p.RouterFactory, func(_ string) (p2p.Router, error){
			return mockRouter,nil
		})

		defer guard.Unpatch()

		txChan :=make(chan core.Request,1)
		//设置第一个通道txchan
		type Request struct {
			Time    int64
			From    string
			To      string
			ReqType int
			Body    []byte
		}
		//构造测试数据
		txChan<-core.Request{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:1,
			Body:[]byte{'a','b'}}

		mockRouter.EXPECT().FilteredChan(nil).Return(txChan,nil)

		//设置第二个通道Blockchan
		blockChan :=make(chan core.Request,1)
		//设置第一个通道txchan
		//构造测试数据
		blockChan<-core.Request{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:2,
			Body:[]byte{'c','d'}}
		mockRouter.EXPECT().FilteredChan(nil).Return(blockChan,nil)

		p,_ := NewDPoS(core.Member{ID:"id0", Pubkey:[]byte{}, Seckey:[]byte{}}, nil)

		p.Genesis(core.Timestamp{},[]byte{})
	})


}

func TestDPoS_Run(t *testing.T) {
	Convey("Test fo Run", t, func() {
		mockCtr :=NewController(t)
		mockRouter:=mock_p2p.NewMockRouter(mockCtr)

		//获取router实例
		guard:= Patch(p2p.RouterFactory, func(_ string) (p2p.Router, error){
			return mockRouter,nil
		})

		defer guard.Unpatch()

		txChan :=make(chan core.Request,1)
		//设置第一个通道txchan
		type Request struct {
			Time    int64
			From    string
			To      string
			ReqType int
			Body    []byte
		}
		//构造测试数据
		txChan<-core.Request{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:1,
			Body:[]byte{'a','b'}}

		mockRouter.EXPECT().FilteredChan(nil).Return(txChan,nil)

		//设置第二个通道Blockchan
		blockChan :=make(chan core.Request,1)
		//设置第一个通道txchan
		//构造测试数据
		blockChan<-core.Request{
			Time:20180426111111,
			From:"0xaaaaaaaaaaaaa",
			To:"0xbbbbbbbbbbbb",
			ReqType:2,
			Body:[]byte{'c','d'}}
		mockRouter.EXPECT().FilteredChan(nil).Return(blockChan,nil)

		p,_ := NewDPoS(core.Member{ID:"id0", Pubkey:[]byte{}, Seckey:[]byte{}}, nil)

		p.Run()

		time.Sleep(20*time.Second)
	})


}