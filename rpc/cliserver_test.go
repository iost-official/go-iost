package rpc

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/mocks"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	//"github.com/iost-official/prototype/network"
	//"github.com/iost-official/prototype/network/mocks"
	"testing"

	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRpcServer(t *testing.T) {
	Convey("Test of RpcServer", t, func() {
		txDb := tx.TxDbInstance()
		So(txDb, ShouldNotBeNil)
		main := lua.NewMethod(vm.Public, "main", 0, 1)
		code := `function main()
			 		Put("hello", "world")
					return "success"
				end`
		lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

		_tx := tx.NewTx(int64(0), &lc)
		acc, _ := account.NewAccount(nil)
		a1, _ := account.NewAccount(nil)
		sig1, _ := tx.SignContract(_tx, a1)
		_tx, _ = tx.SignTx(_tx, acc, sig1)
		Convey("Test of GetTransaction", func() {
			txdb := tx.TxDb
			fmt.Println(txdb)
			err := txdb.Add(&_tx)
			So(err, ShouldBeNil)

			txkey := TransactionKey{
				Publisher: _tx.Publisher.Pubkey,
				Nonce:     _tx.Nonce,
			}
			fmt.Println(txkey.Publisher)
			hs := new(RpcServer)
			_, err = hs.GetTransaction(context.Background(), &txkey)
			So(err, ShouldBeNil)
		})
		//tmp test,better to create new state,insert to StdPool and test it
		Convey("Test of GetState", func() {
			ctl := gomock.NewController(t)
			mockPool := core_mock.NewMockPool(ctl)
			mockPool.EXPECT().Get(gomock.Any()).AnyTimes().Return(state.MakeVString("hello"), nil)
			state.StdPool = mockPool

			hs := new(RpcServer)
			_, err := hs.GetState(context.Background(), &Key{S: "HowHsu"})
			So(err, ShouldBeNil)
		})
		Convey("Test of GetBlock", func() {
			ctl := gomock.NewController(t)
			mockChain := core_mock.NewMockChain(ctl)
			mockChain.EXPECT().Length().AnyTimes().Return(uint64(100))
			mockChain.EXPECT().GetBlockByNumber(gomock.Any()).AnyTimes().Return(&block.Block{
				Head: block.BlockHead{
					Version:    2,
					ParentHash: []byte("parent Hash"),
					TreeHash:   []byte("tree hash"),
					Info:       []byte("info "),
					Number:     int64(0),
					Witness:    "id2,id3,id5,id6",
					Signature:  []byte("Signatrue"),
					Time:       201222,
				},
			})
			block.BChain = mockChain

			hs := new(RpcServer)
			_, err := hs.GetBlock(context.Background(), &BlockKey{Layer: 10})
			So(err, ShouldBeNil)
		})

		Convey("Test of GetBalance", func() {
			ctl := gomock.NewController(t)
			mockPool := core_mock.NewMockPool(ctl)
			mockPool.EXPECT().GetHM(gomock.Any(), gomock.Any()).AnyTimes().Return(state.MakeVFloat(18.0), nil)
			state.StdPool = mockPool

			hs := new(RpcServer)
			balance, err := hs.GetBalance(context.Background(), &Key{S: "HowHsu"})
			So(err, ShouldBeNil)

			vf := state.MakeVFloat(18.0)
			So(balance.Sv, ShouldEqual, vf.EncodeString())

		})

	})
}
