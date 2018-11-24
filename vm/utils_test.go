package vm

import (
	"os"
	"testing"

	"encoding/json"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/native"
)

var testID = []string{
	"IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE",
	"IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY", "8dJ9YKovJ5E7hkebAQaScaG1BA8snRUHPUbVcArcTVq6",
	"IOST7ZNDWeh8pHytAZdpgvp7vMpjZSSe5mUUKxDm6AXPsbdgDMAYhs", "7CnwT7BXkEFAVx6QZqC7gkDhQwbvC3d2CkMZvXHZdDMN",
	"IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd", "Htarc5Sp4trjqY4WrTLtZ85CF6qx87v7CRwtV4RRGnbF",
	"IOST7GmPn8xC1RESMRS6a62RmBcCdwKbKvk2ZpxZpcXdUPoJdapnnh", "Bk8bAyG4VLBcrsoRErPuQGhwCy4C1VxfKE4jjX9oLhv",
	"IOST7ZGQL4k85v4wAxWngmow7JcX4QFQ4mtLNjgvRrEnEuCkGSBEHN", "546aCDG9igGgZqVZeybajaorP5ZeF9ghLu2oLncXk3d6",
	"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9", "DXNYRwG7dRFkbWzMNEbKfBhuS8Yn51x9J6XuTdNwB11M",
	"IOST8mFxe4kq9XciDtURFZJ8E76B8UssBgRVFA5gZN9HF5kLUVZ1BB", "AG8uECmAwFis8uxTdWqcgGD9tGDwoP6CxqhkhpuCdSeC",
	"IOST7uqa5UQPVT9ongTv6KmqDYKdVYSx4DV2reui4nuC5mm5vBt3D9", "GJt5WSSv5WZi1axd3qkb1vLEfxCEgKGupcXf45b5tERU",
	"IOST6wYBsLZmzJv22FmHAYBBsTzmV1p1mtHQwkTK9AjCH9Tg5Le4i4", "7U3uwEeGc2TF3Xde2oT66eTx1Uw15qRqYuTnMd3NNjai",
}

var systemContract = native.SystemABI()

func ininit(t *testing.T) (*database.Visitor, db.MVCCDB) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}
	//mvccdb := replaceDB(t)
	vi := database.NewVisitor(0, mvccdb)
	vi.SetTokenBalance("iost", testID[0], 1000000)
	vi.SetContract(systemContract)
	vi.Commit()
	return vi, mvccdb
}

func closeMVCCDB(m db.MVCCDB) {
	m.Close()
	os.RemoveAll("mvcc")
}

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
	mock.EXPECT().Get("state", "m-auth.iost-auth-a").Return("s"+string(ax), nil)

	b := account.NewInitAccount("b", testID[2], testID[2])
	bx, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	mock.EXPECT().Get("state", "m-auth.iost-auth-b").Return("s"+string(bx), nil)

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
	mock.EXPECT().Get("state", "m-auth.iost-auth-a").AnyTimes().Return("s"+string(ax), nil)

	b := account.NewInitAccount("b", testID[2], testID[2])
	bx, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	mock.EXPECT().Get("state", "m-auth.iost-auth-b").AnyTimes().Return("s"+string(bx), nil)

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
