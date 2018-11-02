package integration

import (
	"fmt"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	. "github.com/iost-official/go-iost/verifier"
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

type fataler interface {
	Fatal(args ...interface{})
}

func prepareContract(t fataler, s *Simulator) {
	kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 18; i += 2 {
		s.SetAccount(account.NewInitAccount(testID[i], testID[i], testID[i]))
		s.Visitor.SetBalance(testID[i], 1000000000)
		s.SetGas(testID[i], 100000)
	}
	// deploy iost.token
	s.SetContract(native.TokenABI())

	// create token
	r, err := s.Call("iost.token", "create", fmt.Sprintf(`["%v", "%v", %v, {}]`, "iost", testID[0], 1000), kp.ID, kp)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}
	// issue token
	r, err = s.Call("iost.token", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", testID[0], "1000"), kp.ID, kp)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}
	if 1e11 != s.Visitor.TokenBalance("iost", testID[0]) {
		t.Fatal(s.Visitor.TokenBalance("iost", testID[0]))
	}
	s.Visitor.Commit()
}

func prepareAuth(t fataler, s *Simulator) *account.KeyPair {
	kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}
	s.SetAccount(account.NewInitAccount(kp.ID, kp.ID, kp.ID))
	return kp
}

var bh = &block.BlockHead{
	ParentHash: []byte("abc"),
	Number:     200,
	Witness:    "witness",
	Time:       123456,
}
