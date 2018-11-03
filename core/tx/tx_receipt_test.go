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
			tx.SuccActionNum = 99
			tx.GasUsage = 88
			tx.Status = Status{
				Code:    ErrorGasRunOut,
				Message: "error gas run out",
			}
			tx.Receipts = append(tx.Receipts, Receipt{
				Type:    SystemDefined,
				Content: "{\"num\": 1, \"message\": \"contract1\"}",
			})
			tx1 := NewTxReceipt([]byte{})

			encode := tx.Encode()
			err := tx1.Decode(encode)
			So(err, ShouldEqual, nil)

			hash := tx.Hash()
			hash1 := tx1.Hash()
			So(bytes.Equal(hash, hash1), ShouldEqual, true)

			So(bytes.Equal(tx.TxHash, tx1.TxHash), ShouldBeTrue)
			So(tx.SuccActionNum == tx1.SuccActionNum, ShouldBeTrue)
			So(tx.GasUsage == tx1.GasUsage, ShouldBeTrue)
			So(tx.Status.Code == tx1.Status.Code, ShouldBeTrue)
			So(tx.Status.Message == tx1.Status.Message, ShouldBeTrue)
			So(len(tx.Receipts) == len(tx1.Receipts), ShouldBeTrue)
			for i := 0; i < len(tx.Receipts); i++ {
				So(tx.Receipts[i].Type == tx1.Receipts[i].Type, ShouldBeTrue)
				So(tx.Receipts[i].Content == tx1.Receipts[i].Content, ShouldBeTrue)
			}

		})

	})
}
