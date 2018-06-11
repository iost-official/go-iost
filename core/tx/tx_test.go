package tx

import (
	"testing"
	//	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"

	"github.com/iost-official/prototype/vm/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func gentx() Tx {
	main := lua.NewMethod(0, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("ahaha")}, code, main)

	return NewTx(int64(0), &lc)
}
func TestTx(t *testing.T) {
	Convey("Test of Tx", t, func() {
		ctl := gomock.NewController(t)

		mockContract := vm_mock.NewMockContract(ctl)
		mockContract.EXPECT().Encode().AnyTimes().Return([]byte{1, 2, 3})

		a1, _ := account.NewAccount(nil)
		a2, _ := account.NewAccount(nil)
		a3, _ := account.NewAccount(nil)
		/*
			Convey("string", func() {
			 	tx := gentx()
				fmt.Println(tx.String())
				//So(err.Error(), ShouldEqual, "signer error")
			})
		*/
		Convey("sign and verify", func() {
			tx := NewTx(int64(0), mockContract)
			sig1, err := SignContract(tx, a1)

			So(tx.VerifySigner(sig1), ShouldBeTrue)

			tx.Signs = append(tx.Signs, sig1)
			sig2, err := SignContract(tx, a2)
			tx.Signs = append(tx.Signs, sig2)

			err = tx.VerifySelf()
			So(err.Error(), ShouldEqual, "publisher error")

			tx3, err := SignTx(tx, a3)
			So(err, ShouldBeNil)
			err = tx3.VerifySelf()
			So(err, ShouldBeNil)

			tx.Signs[0] = common.Signature{
				Algorithm: common.Secp256k1,
				Sig:       []byte("hello"),
				Pubkey:    []byte("world"),
			}
			err = tx.VerifySelf()
			So(err.Error(), ShouldEqual, "signer error")
		})

	})
}
