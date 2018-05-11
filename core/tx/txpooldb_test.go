package tx

import (
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/vm/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTxPoolDb(t *testing.T) {
	Convey("Test of TestTxPoolDb", t, func() {
		ctl := gomock.NewController(t)

		mockContract := vm_mock.NewMockContract(ctl)
		mockContract.EXPECT().Encode().AnyTimes().Return([]byte{1, 2, 3, 4})

		Convey("Test of Add", func() {
			txpooldb, err := NewTxPoolDbImpl()
			tx := NewTx(int64(0), mockContract)

			err = txpooldb.Add(tx)
			So(err, ShouldBeNil)
			txpooldb.Close()
		})
		Convey("Test of Get", func() {
			txpooldb, err := NewTxPoolDbImpl()
			tx := NewTx(int64(0), mockContract)
			err = txpooldb.Add(tx)
			hash := tx.Hash()

			_, err = txpooldb.Get(hash)
			So(err, ShouldBeNil)
			//todo: test *txPtr==tx
			txpooldb.Close()
		})
	})
}
