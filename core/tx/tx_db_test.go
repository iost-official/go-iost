package tx

import (
	"testing"

	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
)

func genTx() Tx {
	main := lua.NewMethod(0, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

	return NewTx(int64(0), &lc)
}
func TestTxDb(t *testing.T) {
	Convey("Test of TxDb", t, func() {
		txdb := TxDbInstance()
		Convey("Test of Add", func() {
			_tx := genTx()
			err := txdb.Add(&_tx)
			So(err, ShouldBeNil)
		})

		Convey("Test of Has", func() {
			_tx := genTx()
			_, err := txdb.Has(&_tx)
			So(err, ShouldBeNil)
			txdb.Add(&_tx)

			_, err = txdb.Has(&_tx)
			So(err, ShouldBeNil)
		})

		Convey("Test of Get", func() {
			tx1 := genTx()
			err := txdb.Add(&tx1)
			hash := tx1.Hash()
			tx2, err := txdb.Get(hash)
			So(err, ShouldBeNil)
			So(tx2.Time, ShouldEqual, tx1.Time)
		})

		Convey("Test of GetByPN", func() {
			tx1 := genTx()
			err := txdb.Add(&tx1)
			tx2, err := txdb.(*TxPoolDb).GetByPN(tx1.Nonce, tx1.Publisher.Pubkey)
			So(err, ShouldBeNil)
			So(tx1.Time, ShouldEqual, (*tx2).Time)
		})
	})
}
