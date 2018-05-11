package tx

import (
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTxPoolDb(t *testing.T) {
	Convey("Test of TestTxPoolDb", t, func() {

		Convey("Test of Add", func() {
			txpooldb, err := NewTxPoolDbImpl()

			main := lua.NewMethod("main", 0, 1)
			code := `function main()
						Put("hello", "world")
						return "success"
					end`
			lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Sender: vm.IOSTAccount("ahaha")}, code, main)

			tx := NewTx(int64(0), &lc)
			err = txpooldb.Add(&tx)
			So(err, ShouldBeNil)
			txpooldb.(*TxPoolDbImpl).Close()
		})

		Convey("Test of Get", func() {
			txpooldb, err := NewTxPoolDbImpl()
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
			//todo: test *txPtr==tx
			txpooldb.(*TxPoolDbImpl).Close()
		})

	})
}
