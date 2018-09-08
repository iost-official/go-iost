package tx

import (
	"testing"

	"os"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	. "github.com/smartystreets/goconvey/convey"
)

/*
func genTx() Tx {
	main := lua.NewMethod(0, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

	return NewTx(int64(0), &lc)
}
*/

func TestTxDb(t *testing.T) {
	Convey("Test of TxDb", t, func() {
		txDb, _ := NewTxDB("./txDB/")
		a1, _ := account.NewAccount(nil, crypto.Secp256k1)
		tx1 := NewTx([]*Action{}, [][]byte{a1.Pubkey}, 100000, 100, 11)
		tx2 := NewTx([]*Action{}, [][]byte{a1.Pubkey}, 88888, 22, 11)
		var txs []*Tx
		txs = make([]*Tx, 0)
		txs = append(txs, tx1)

		re1 := NewTxReceipt(tx1.Hash())
		re2 := NewTxReceipt(tx2.Hash())

		var res []*TxReceipt
		res = make([]*TxReceipt, 0)
		res = append(res, &re1)

		err := txDb.Push(txs, res)
		So(err, ShouldBeNil)

		Convey("Test of HasTx", func() {
			b, err := txDb.HasTx(tx1.Hash())
			So(b, ShouldBeTrue)
			So(err, ShouldBeNil)

			b, err = txDb.HasTx(tx2.Hash())
			So(b, ShouldBeFalse)
			So(err, ShouldBeNil)

		})

		Convey("Test of GetTx", func() {
			tx, err := txDb.GetTx(tx1.Hash())
			So(err, ShouldBeNil)
			So(string(tx.Hash()), ShouldEqual, string(tx1.Hash()))

			_, err = txDb.GetTx(tx2.Hash())
			So(err, ShouldNotBeNil)

		})

		Convey("Test of GetReceipt", func() {
			re, err := txDb.GetReceipt(re1.Hash())
			So(err, ShouldBeNil)
			So(string(re.Hash()), ShouldEqual, string(re1.Hash()))

			_, err = txDb.GetTx(re2.Hash())
			So(err, ShouldNotBeNil)
		})

		Convey("Test of GetReceiptByTxHash", func() {
			re, err := txDb.GetReceiptByTxHash(tx1.Hash())
			So(err, ShouldBeNil)
			So(string(re.Hash()), ShouldEqual, string(re1.Hash()))

			_, err = txDb.GetReceiptByTxHash(tx2.Hash())
			So(err, ShouldNotBeNil)
		})

		Convey("Test of HasReceipt", func() {
			b, err := txDb.HasReceipt(re1.Hash())
			So(err, ShouldBeNil)
			So(b, ShouldBeTrue)

			b, err = txDb.HasReceipt(re2.Hash())
			So(err, ShouldBeNil)
			So(b, ShouldBeFalse)
		})

	})

	os.RemoveAll("txDB")
}
