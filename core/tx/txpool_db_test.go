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
			dbtxpool, err := TxPoolFactory("db")

			main := lua.NewMethod("main", 0, 1)
			code := `function main()
						Put("hello", "world")
						return "success"
					end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

			tx := NewTx(int64(0), &lc)
			err = dbtxpool.Add(&tx)
			So(err, ShouldBeNil)
			dbtxpool.(*TxPoolDb).Close()
		})

		Convey("Test of Get", func() {
			dbtxpool, err := NewTxPoolDb()
			main := lua.NewMethod("main", 0, 1)
			code := `function main()
							Put("hello", "world")
							return "success"
						end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

			tx := NewTx(int64(0), &lc)
			err = dbtxpool.Add(&tx)
			hash := tx.Hash()

			tx1, err := dbtxpool.Get(hash)
			So(err, ShouldBeNil)
			So(tx.Time, ShouldEqual, (*tx1).Time)
			dbtxpool.(*TxPoolDb).Close()
		})

	})
}
