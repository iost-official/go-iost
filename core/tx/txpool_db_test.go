package tx

import (
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTxPoolDb(t *testing.T) {
	Convey("Test of TxPoolDb", t, func() {

		Convey("Test of Add", func() {
			txpooldb, err := TxPoolFactory("db")

			main := lua.NewMethod("main", 0, 1)
			code := `function main()
						Put("hello", "world")
						return "success"
					end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

			tx := NewTx(int64(0), &lc)
			err = txpooldb.Add(&tx)
			So(err, ShouldBeNil)
			//txpooldb.(*TxPoolDb).Close()
		})

		Convey("Test of Has", func() {
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

		Convey("Test of Get", func() {
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

	})
}
