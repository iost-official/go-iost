package verifier

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	. "github.com/smartystreets/goconvey/convey"
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

var bh = &block.BlockHead{
	ParentHash: []byte("abc"),
	Number:     200,
	Witness:    "witness",
	Time:       123456,
}

func TestTransfer(t *testing.T) { // todo auth error
	ilog.Stop()

	s := NewSimulator()
	defer s.Clear()
	s.Visitor.SetBalance(testID[0], 10000000)

	kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}

	r, err := s.Call("iost.system", "Transfer", fmt.Sprintf(`["%v","%v","%v"]`, testID[0], testID[2], 0.0001), kp)

	Convey("test transfer", t, func() {
		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
		So(s.Visitor.Balance(testID[0]), ShouldEqual, int64(9990000))
		So(s.Visitor.Balance(testID[2]), ShouldEqual, int64(10000))
	})
}

func TestJS_Database(t *testing.T) {
	//ilog.Stop()

	s := NewSimulator()
	defer s.Clear()
	s.Visitor.SetBalance(testID[0], 10000000)

	c, err := s.Compile("datatbase", "test_data/database", "test_data/database")
	if err != nil {
		t.Fatal(err)
	}

	kp, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}

	cname := s.DeployContract(c, kp)
	t.Log("cname ", cname)

	Convey("test of s database", t, func() {
		So(s.Visitor.Contract(cname), ShouldNotBeNil)
		So(s.Visitor.Get(cname+"-"+"num"), ShouldEqual, "s9")
		So(s.Visitor.Get(cname+"-"+"string"), ShouldEqual, "shello")
		So(s.Visitor.Get(cname+"-"+"bool"), ShouldEqual, "strue")
		So(s.Visitor.Get(cname+"-"+"array"), ShouldEqual, "s[1,2,3]")
		So(s.Visitor.Get(cname+"-"+"obj"), ShouldEqual, `s{"foo":"bar"}`)

		r, err := s.Call(cname, "read", `[]`, kp)

		So(err, ShouldBeNil)
		So(r.Status.Message, ShouldEqual, "")
	})

}

/*

func TestGenesis(t *testing.T) {
	ilog.Stop()
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("mvcc")

	var acts []*tx.Action
	for i := 0; i < 3; i++ {
		act := tx.NewAction("iost.system", "IssueIOST", fmt.Sprintf(`["%v", %v]`, testID[2*i], "1000000000000000"))
		acts = append(acts, &act)
	}
	// deploy iost.vote
	voteFilePath := "../contract/vote.js"
	voteAbiPath := "../contract/vote.js.abi"
	fd, err := ioutil.ReadFile(voteFilePath)
	if err != nil {
		t.Fatal(err)
	}
	rawCode := string(fd)
	fd, err = ioutil.ReadFile(voteAbiPath)
	if err != nil {
		t.Fatal(err)
	}
	rawAbi := string(fd)
	c := contract.Compiler{}
	code, err := c.Parse("iost.vote", rawCode, rawAbi)
	if err != nil {
		t.Fatal(err)
	}

	act := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.vote", code.B64Encode()))
	acts = append(acts, &act)

	num := 3
	for i := 0; i < num; i++ {
		act1 := tx.NewAction("iost.vote", "InitProducer", fmt.Sprintf(`["%v"]`, testID[2*i]))
		acts = append(acts, &act1)
	}
	act11 := tx.NewAction("iost.vote", "InitAdmin", fmt.Sprintf(`["%v"]`, testID[0]))
	acts = append(acts, &act11)

	// deploy iost.bonus
	act2 := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.bonus", native.BonusABI().B64Encode()))
	acts = append(acts, &act2)

	trx := tx.NewTx(acts, nil, 100000000, 0, 0)
	trx.Time = 0
	acc, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}
	trx, err = tx.SignTx(trx, acc)
	if err != nil {
		t.Fatal(err)
	}
	blockHead := block.BlockHead{
		Version:    0,
		ParentHash: nil,
		Number:     0,
		Witness:    acc.ID,
		Time:       0,
	}
	v := Verifier{}
	txr, err := v.Exec(&blockHead, mvccdb, trx, time.Millisecond*100)
	if err != nil || txr.Status.Code != tx.Success {
		t.Fatal(err, txr)
	}
	fmt.Println(txr)

	vi := database.NewVisitor(0, mvccdb)
	fmt.Println(vi.Get("iost.vote-" + "pendingBlockNumber"))
	fmt.Println(vi.Balance(testID[0]))
}

func TestDomain(t *testing.T) {
	js := NewJSTester(t)
	defer js.Clear()

	lc, err := ReadFile("../vm/test_data/database.js")
	if err != nil {
		t.Fatal(err)
	}
	js.SetJS(string(lc))
	js.SetAPI("read")
	js.SetAPI("change")
	js.DoSet()
	//t.Log("========= constructor")
	Convey("test of js database", t, func() {
		So(js.ReadDB("num").(string), ShouldEqual, "9")
		So(js.ReadDB("string").(string), ShouldEqual, "hello")
		So(js.ReadDB("bool").(string), ShouldEqual, "true")
		So(js.ReadDB("array").(string), ShouldEqual, "[1,2,3]")
		So(js.ReadDB("obj").(string), ShouldEqual, `{"foo":"bar"}`)
	})
	js.vi.SetContract(native.ABI("iost.domain", native.DomainABIs))
	js.vi.Commit()
	js.Call("iost.domain", "Link", fmt.Sprintf(`["abcde","%v"]`, js.cname))
	js.Call("abcde", "read", "[]")

}

func array2json(ss []interface{}) string {
	x, err := json.Marshal(ss)
	if err != nil {
		panic(err)
	}
	return string(x)
}

func TestAuthority(t *testing.T) {
	js := NewJSTester(t)
	defer js.Clear()

	ca, err := Compile("iost.auth", "../contract/account", "../contract/account.js")
	if err != nil {
		t.Fatal(err)
	}
	js.vi.SetContract(ca)
	js.vi.Commit()
	js.cname = "iost.auth"
	Convey("test of Auth", t, func() {
		js.Call("iost.auth", "SignUp", array2json([]interface{}{"myid", "okey", "akey"}))
		So(js.ReadMap("account", "myid"), ShouldEqual, `{"id":"myid","permissions":{"active":{"name":"active","groups":[],"items":[{"id":"akey","is_key_pair":true,"weight":1}],"threshold":1},"owner":{"name":"owner","groups":[],"items":[{"id":"okey","is_key_pair":true,"weight":1}],"threshold":1}}}`)

		js.Call("iost.auth", "AddPermission", array2json([]interface{}{"myid", "perm1", 1}))
		So(js.ReadMap("account", "myid"), ShouldContainSubstring, `"perm1":{"name":"perm1","groups":[],"items":[],"threshold":1}`)
	})

}
*/

func prepareContract(t *testing.T, js *JSTester) {
	bh = &block.BlockHead{
		ParentHash: []byte("abc"),
		Number:     0,
		Witness:    "witness",
		Time:       123456,
	}
	for i := 0; i < 18; i++ {
		js.vi.MPut("iost.auth-account", testID[i], database.MustMarshal(fmt.Sprintf(`{"id":"%s","permissions":{"active":{"name":"active","groups":[],"items":[{"id":"%s","is_key_pair":true,"weight":1}],"threshold":1},"owner":{"name":"owner","groups":[],"items":[{"id":"%s","is_key_pair":true,"weight":1}],"threshold":1}}}`, testID[i], testID[i], testID[i])))
	}
	js.vi.Commit()
	// deploy iost.token
	r := js.Call("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.token", native.TokenABI().B64Encode()))
	if r.Status.Code != tx.Success {
		t.Fatal(r)
	}
	// create token
	r = js.Call("iost.token", "create", fmt.Sprintf(`["%v", "%v", %v, {}]`, "iost", testID[0], 1000))
	if r.Status.Code != tx.Success {
		t.Fatal(r)
	}
	// issue token
	r = js.Call("iost.token", "issue", fmt.Sprintf(`["%v", "%v", "%v"]`, "iost", testID[0], "1000"))
	if r.Status.Code != tx.Success {
		t.Fatal(r)
	}
	if 1e11 != js.vi.TokenBalance("iost", testID[0]) {
		t.Fatal(js.vi.TokenBalance("iost", testID[0]))
	}
	js.vi.Commit()
}

func TestAmountLimit(t *testing.T) {
	ilog.Stop()
	Convey("test of amount limit", t, func() {
		js := NewJSTester(t)
		defer js.Clear()
		prepareContract(t, js)

		ca, err := Compile("Contracttransfer", "./test_data/transfer", "./test_data/transfer.js")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		js.vi.SetContract(ca)
		js.vi.Commit()

		ca, err = Compile("Contracttransfer1", "./test_data/transfer1", "./test_data/transfer1.js")
		if err != nil || ca == nil {
			t.Fatal(err)
		}
		js.vi.SetContract(ca)
		js.vi.Commit()
		js.cname = "Contracttransfer1"

		Reset(func() {
			js.vi.SetTokenBalanceFixed("iost", testID[0], "1000")
			js.vi.SetTokenBalanceFixed("iost", testID[2], "0")
		})

		Convey("test of amount limit", func() {
			r := js.Call("Contracttransfer", "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "10"))
			js.vi.Commit()
			So(r.Status.Code, ShouldEqual, tx.Success)
			balance0 := common.Fixed{Value:js.vi.TokenBalance("iost", testID[0]), Decimal:js.vi.Decimal("iost")}
			balance2 := common.Fixed{Value:js.vi.TokenBalance("iost", testID[2]), Decimal:js.vi.Decimal("iost")}
			So(balance0.ToString(), ShouldEqual, "990")
			So(balance2.ToString(), ShouldEqual, "10")
		})

		Convey("test out of amount limit", func() {
			r := js.Call("Contracttransfer", "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "110"))
			js.vi.Commit()
			So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
			So(r.Status.Message, ShouldContainSubstring, "exceed amountLimit in abi")
			//balance0 := common.Fixed{Value:js.vi.TokenBalance("iost", testID[0]), Decimal:js.vi.Decimal("iost")}
			//balance2 := common.Fixed{Value:js.vi.TokenBalance("iost", testID[2]), Decimal:js.vi.Decimal("iost")}
			// todo exit when monitor.Call return err
			// So(balance0.ToString(), ShouldEqual, "990")
			// So(balance2.ToString(), ShouldEqual, "10")
		})

		Convey("test amount limit two level invocation", func() {
			r := js.Call("Contracttransfer1", "transfer", fmt.Sprintf(`["%v", "%v", "%v"]`, testID[0], testID[2], "120"))
			js.vi.Commit()
			So(r.Status.Code, ShouldEqual, tx.Success)
			balance0 := common.Fixed{Value:js.vi.TokenBalance("iost", testID[0]), Decimal:js.vi.Decimal("iost")}
			balance2 := common.Fixed{Value:js.vi.TokenBalance("iost", testID[2]), Decimal:js.vi.Decimal("iost")}
			So(balance0.ToString(), ShouldEqual, "880")
			So(balance2.ToString(), ShouldEqual, "120")
		})

	})
}
