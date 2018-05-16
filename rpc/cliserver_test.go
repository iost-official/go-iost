package rpc

import (
	"context"

	//"github.com/golang/mock/gomock"
	//"github.com/iost-official/prototype/rpc/mock_rpc"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHttpServer(t *testing.T) {
	Convey("Test of HttpServer", t, func() {
		/*
			ctl := gomock.NewController(t)
			mockCtx := rpc_mock.mockContext(ctl)
		*/
		Convey("Test of PublishTx", func() {

			main := lua.NewMethod("main", 0, 1)
			code := `function main()
						Put("hello", "world")
						return "success"
					end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

			tx := tx.NewTx(int64(0), &lc)
			txpb := Transaction{Tx: tx.Encode()}
			hs := new(HttpServer)
			res, err := hs.PublishTx(context.Background(), &txpb)
			So(err, ShouldBeNil)
			So(res.Code, ShouldEqual, 0)
		})
		/*
			Convey("Test of GetTransaction", func() {
				txpooldb, err := TxPoolFactory("db")

				main := lua.NewMethod("main", 0, 1)
				code := `function main()
							Put("hello", "world")
							return "success"
						end`
				lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

				tx := NewTx(int64(0), &lc)
				_, err = txpooldb.Has(&tx)
				So(err, ShouldBeNil)
				txpooldb.Add(&tx)

				_, err = txpooldb.Has(&tx)
				So(err, ShouldBeNil)
				//txpooldb.(*TxPoolDb).Close()
			})

			Convey("Test of GetBalance", func() {
				txpooldb, err := NewTxPoolDb()
				main := lua.NewMethod("main", 0, 1)
				code := `function main()
								Put("hello", "world")
								return "success"
							end`
				lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

				tx := NewTx(int64(0), &lc)
				err = txpooldb.Add(&tx)
				hash := tx.Hash()

				tx1, err := txpooldb.Get(hash)
				So(err, ShouldBeNil)
				So(tx.Time, ShouldEqual, (*tx1).Time)
				//txpooldb.(*TxPoolDb).Close()
			})

			Convey("Test of GetState", func() {
				txpooldb, err := NewTxPoolDb()
				main := lua.NewMethod("main", 0, 1)
				code := `function main()
								Put("hello", "world")
								return "success"
							end`
				lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

				tx := NewTx(int64(0), &lc)
				err = txpooldb.Add(&tx)
				tx1, err := txpooldb.(*TxPoolDb).GetByPN(tx.Nonce,tx.Publisher)
				So(err, ShouldBeNil)
				So(tx.Time, ShouldEqual, (*tx1).Time)
			})

			Convey("Test of GetBlock", func() {
				txpooldb, err := NewTxPoolDb()
				main := lua.NewMethod("main", 0, 1)
				code := `function main()
								Put("hello", "world")
								return "success"
							end`
				lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

				tx := NewTx(int64(0), &lc)
				err = txpooldb.Add(&tx)
				tx1, err := txpooldb.(*TxPoolDb).GetByPN(tx.Nonce,tx.Publisher)
				So(err, ShouldBeNil)
				So(tx.Time, ShouldEqual, (*tx1).Time)
			})
		*/
	})
}
