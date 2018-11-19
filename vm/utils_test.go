package vm

import (
	"testing"

	"encoding/json"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/vm/database"
)

func TestCheckPublisher(t *testing.T) {
	tr := tx.NewTx([]*tx.Action{{
		"system.iost",
		"Transfer",
		"[]",
	}}, []string{}, 10000, 1, 10000, 0)

	kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}
	t2, err := tx.SignTx(tr, "a", []*account.KeyPair{kp})
	if err != nil {
		t.Fatal(err)
	}

	ctl := gomock.NewController(t)

	a := account.NewInitAccount("a", testID[0], testID[0])
	ax, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	mock := database.NewMockIMultiValue(ctl)
	mock.EXPECT().Get("state", "m-auth.iost-account-a").Return("s"+string(ax), nil)

	b := account.NewInitAccount("b", testID[2], testID[2])
	bx, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	mock.EXPECT().Get("state", "m-auth.iost-account-b").Return("s"+string(bx), nil)

	err = CheckPublisher(mock, t2)
	if err != nil {
		t.Fatal(err)
	}

	t2.Publisher = "b"
	err = CheckPublisher(mock, t2)
	if err == nil {
		t.Fatal(err)
	}
}

func TestCheckSigners(t *testing.T) {
	ctl := gomock.NewController(t)
	mock := database.NewMockIMultiValue(ctl)

	a := account.NewInitAccount("a", testID[0], testID[0])
	ax, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	mock.EXPECT().Get("state", "m-auth.iost-account-a").AnyTimes().Return("s"+string(ax), nil)

	b := account.NewInitAccount("b", testID[2], testID[2])
	bx, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	mock.EXPECT().Get("state", "m-auth.iost-account-b").AnyTimes().Return("s"+string(bx), nil)

	kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}
	kp2, err := account.NewKeyPair(common.Base58Decode(testID[3]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}

	tr := tx.NewTx([]*tx.Action{{
		"system.iost",
		"Transfer",
		"[]",
	}}, []string{"a@acitve", "b@acitve"}, 10000, 1, 10000, 0)

	sig1, err := tx.SignTxContent(tr, "a", kp)
	if err != nil {
		t.Fatal(err)
	}
	tr.Signs = append(tr.Signs, sig1)

	err = CheckSigners(mock, tr)
	if err == nil {
		t.Fatal(err)
	}

	sig2, err := tx.SignTxContent(tr, "b", kp2)
	if err != nil {
		t.Fatal(err)
	}
	tr.Signs = append(tr.Signs, sig2)

	err = CheckSigners(mock, tr)
	if err != nil {
		t.Fatal(err)
	}

}
