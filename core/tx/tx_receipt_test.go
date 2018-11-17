package tx

import (
	"testing"

	"bytes"

	. "github.com/smartystreets/goconvey/convey"
)

/*
func gentx() Tx {
	main := lua.NewMethod(0, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, PublishSign: vm.IOSTAccount("ahaha")}, code, main)

	return NewTx(int64(0), &lc, [1]byte)
}
*/
func TestTxReceipt(t *testing.T) {
	Convey("Test of Tx Receipt", t, func() {

		Convey("encode and decode", func() {
			tx := NewTxReceipt([]byte{0, 1, 2})
			tx.Returns = []string{"0"}

			tx.GasUsage = 88
			tx.Status = &Status{
				Code:    ErrorGasRunOut,
				Message: "error gas run out",
			}
			tx.Receipts = append(tx.Receipts, &Receipt{
				FuncName: "iost.token/transfer",
				Content:  "{\"num\": 1, \"message\": \"contract1\"}",
			})
			tx.RAMUsage = map[string]int64{
				"aaa": 1111,
				"bbb": 2222,
				"ccc": 333,
			}
			tx1 := NewTxReceipt([]byte{})

			encode := tx.Encode()
			err := tx1.Decode(encode)
			So(err, ShouldEqual, nil)

			hash := tx.Hash()
			hash1 := tx1.Hash()
			So(bytes.Equal(hash, hash1), ShouldEqual, true)

			So(bytes.Equal(tx.TxHash, tx1.TxHash), ShouldBeTrue)
			So(tx.GasUsage, ShouldEqual, tx1.GasUsage)
			So(tx.Status.Code, ShouldEqual, tx1.Status.Code)
			So(tx.Status.Message, ShouldEqual, tx1.Status.Message)
			So(len(tx.Receipts), ShouldEqual, len(tx1.Receipts))
			for i := 0; i < len(tx.Receipts); i++ {
				So(tx.Returns[i], ShouldEqual, tx1.Returns[i])
				So(tx.Receipts[i].FuncName, ShouldEqual, tx1.Receipts[i].FuncName)
				So(tx.Receipts[i].Content, ShouldEqual, tx1.Receipts[i].Content)
			}

		})

	})
}
