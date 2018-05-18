package tx

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/vm/mocks"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTxPoolImpl(t *testing.T) {
	Convey("Test of TxPool", t, func() {
		var txp TxPool
		txp = NewTxPoolImpl()
		//txp, _ = NewTxPoolStack()
		ctl := gomock.NewController(t)

		mockContract := vm_mock.NewMockContract(ctl)
		mockContract.EXPECT().Encode().AnyTimes().Return([]byte{1, 2, 3})
		tx := NewTx(int64(0), mockContract)
		Convey("Add", func() {
			txp.Add(&tx)
			So(txp.Size(), ShouldEqual, 1)
		})
		Convey("Del", func() {
			txp.Add(&tx)
			So(txp.Size(), ShouldEqual, 1)
			txp.Del(&tx)
			So(txp.Size(), ShouldEqual, 0)
		})

		Convey("Get", func() {
			txp.Add(&tx)
			ttx, ok := txp.Get(tx.Hash())
			So(ok, ShouldBeNil)
			So(string(ttx.Encode()), ShouldEqual, string(tx.Encode()))
			tx2 := NewTx(int64(1), mockContract)
			ttx, ok = txp.Get(tx2.Hash())
			So(ttx, ShouldBeNil)
		})

		Convey("Top", func() {
			txp.Add(&tx)
			ttx, ok := txp.Top()
			So(ok, ShouldBeNil)
			So(string(ttx.Encode()), ShouldEqual, string(tx.Encode()))
			txp.Del(&tx)
			tttx, _ := txp.Top()
			So(tttx, ShouldBeNil)
		})

		Convey("Has", func() {
			txp.Add(&tx)
			bt, err := txp.Has(&tx)
			So(err, ShouldBeNil)
			So(bt, ShouldBeTrue)
			tx2 := NewTx(int64(1), mockContract)
			bt, _ = txp.Has(&tx2)
			So(bt, ShouldBeFalse)
		})

	})
}
