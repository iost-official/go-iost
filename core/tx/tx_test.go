package tx

import (
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/vm/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTx(t *testing.T) {
	Convey("Test of Tx", t, func() {
		ctl := gomock.NewController(t)

		mockContract := vm_mock.NewMockContract(ctl)
		mockContract.EXPECT().Encode().AnyTimes().Return([]byte{1, 2, 3})

		a1, _ := account.NewAccount(nil)
		a2, _ := account.NewAccount(nil)
		a3, _ := account.NewAccount(nil)

		Convey("sign and verify", func() {
			tx := NewTx(int64(0), mockContract)
			tx1, err := SignContract(tx, a1)
			So(err, ShouldBeNil)
			So(len(tx1.Signs), ShouldEqual, 1)
			tx2, err := SignContract(tx1, a2)
			So(err, ShouldBeNil)
			So(len(tx2.Signs), ShouldEqual, 2)

			err = tx2.VerifySelf()
			So(err.Error(), ShouldEqual, "publisher error")

			tx3, err := SignTx(tx2, a3)
			So(err, ShouldBeNil)
			err = tx3.VerifySelf()
			So(err, ShouldBeNil)

			tx1.Signs[0] = common.Signature{common.Secp256k1, []byte("hello"), []byte("world")}
			err = tx1.VerifySelf()
			So(err.Error(), ShouldEqual, "signer error")
		})
	})
}
