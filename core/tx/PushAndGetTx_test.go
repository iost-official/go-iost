package tx

import (
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/vm/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPushAndGetTx(t *testing.T) {
	Convey("Test of PushAndGetTx", t, func() {
		ctl := gomock.NewController(t)

		mockContract := vm_mock.NewMockContract(ctl)
		mockContract.EXPECT().Encode().AnyTimes().Return([]byte{1, 2, 3})
		Convey("Test of PushTx", func() {
			tx := NewTx(int64(0), mockContract)
			err := tx.PushTx()
			So(err, ShouldBeNil)
		})
		Convey("Test of GetTx", func() {
			tx := NewTx(int64(0), mockContract)
			err := tx.PushTx()
			hash := tx.Hash()

			_tx := NewTx(int64(0), mockContract)
			err = _tx.GetTx(hash)
			So(err, ShouldBeNil)
			So(_tx, ShouldNotBeNil)
			//todo: test tx==_tx
		})
	})
}
