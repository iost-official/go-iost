package global

import (
	"testing"

	"os"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
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
		txDb, err := NewTxDB("./txDB/")
		So(err, ShouldBeNil)

		a1, _ := account.NewAccount(nil, crypto.Secp256k1)
		tx1 := tx.NewTx([]*tx.Action{}, [][]byte{a1.Pubkey}, 100000, 100, 11)
		tx2 := tx.NewTx([]*tx.Action{}, [][]byte{a1.Pubkey}, 88888, 22, 11)
		var txs []*tx.Tx
		txs = make([]*tx.Tx, 0)
		txs = append(txs, tx1)

		re1 := tx.NewTxReceipt(tx1.Hash())
		re2 := tx.NewTxReceipt(tx2.Hash())

		var res []*tx.TxReceipt
		res = make([]*tx.TxReceipt, 0)
		res = append(res, &re1)

		err = txDb.Push(txs, res)
		So(err, ShouldBeNil)

		b, err := txDb.HasTx(tx1.Hash())
		So(b, ShouldBeTrue)
		So(err, ShouldBeNil)

		b, err = txDb.HasTx(tx2.Hash())
		So(b, ShouldBeFalse)
		So(err, ShouldBeNil)

		tx, err := txDb.GetTx(tx1.Hash())
		So(err, ShouldBeNil)
		So(string(tx.Hash()), ShouldEqual, string(tx1.Hash()))

		_, err = txDb.GetTx(tx2.Hash())
		So(err, ShouldNotBeNil)

		re, err := txDb.GetReceipt(re1.Hash())
		So(err, ShouldBeNil)
		So(string(re.Hash()), ShouldEqual, string(re1.Hash()))

		_, err = txDb.GetTx(re2.Hash())
		So(err, ShouldNotBeNil)

		re, err = txDb.GetReceiptByTxHash(tx1.Hash())
		So(err, ShouldBeNil)
		So(string(re.Hash()), ShouldEqual, string(re1.Hash()))

		_, err = txDb.GetReceiptByTxHash(tx2.Hash())
		So(err, ShouldNotBeNil)

		b, err = txDb.HasReceipt(re1.Hash())
		So(err, ShouldBeNil)
		So(b, ShouldBeTrue)

		b, err = txDb.HasReceipt(re2.Hash())
		So(err, ShouldBeNil)
		So(b, ShouldBeFalse)

	})

	os.RemoveAll("txDB")
}
