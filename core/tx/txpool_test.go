package tx

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/vm/mocks"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTxPoolImpl(t *testing.T) {
	Convey("Test of TxPool", t, func() {
		txp := NewTxPoolImpl()
		ctl := gomock.NewController(t)

		mockContract := vm_mock.NewMockContract(ctl)
		mockContract.EXPECT().Encode().AnyTimes().Return([]byte{1, 2, 3})
		tx := NewTx(int64(0), mockContract)
		Convey("Add", func() {
			txp.Add(tx)
			So(len(txp.txMap), ShouldEqual, 1)
		})

		Convey("Del", func() {
			txp.Del(tx)
			So(len(txp.txMap), ShouldEqual, 0)
		})

		/*
			Convey("Find", func() {
				txp.Add(tx)
				tx2, err := txp.Find(tx.Hash())
				So(err, ShouldBeNil)
				So(tx2.Time, ShouldEqual, tx.Time)

				_, err = txp.Find([]byte("hello"))
				So(err, ShouldNotBeNil)
			})
		*/

		Convey("Has", func() {
			txp.Add(tx)
			bt, err := txp.Has(tx)
			So(err, ShouldBeNil)
			So(bt, ShouldBeTrue)
			tx2 := NewTx(int64(1), mockContract)
			bt, _ = txp.Has(tx2)
			So(bt, ShouldBeFalse)
		})

		Convey("GetSlice", func() {
			txp.Add(tx)
			s, err := txp.GetSlice()
			So(err, ShouldBeNil)
			So(len(s), ShouldEqual, 1)
			So(s[0].Time, ShouldEqual, tx.Time)
		})

		Convey("Size", func() {
			txp.Add(tx)
			l := txp.Size()
			So(l, ShouldEqual, 1)
		})

	})
}
