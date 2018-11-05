package vm

import (
	"fmt"
	"os"
	"testing"
	"time"

	"strconv"

	"encoding/json"
	"io/ioutil"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/native"
	. "github.com/smartystreets/goconvey/convey"
)

func watchTime(f func()) time.Duration {
	ta := time.Now()
	f()
	return time.Now().Sub(ta)
}

func Compile(id, src, abi string) (*contract.Contract, error) {
	bs, err := ioutil.ReadFile(src + ".js")
	if err != nil {
		return nil, err
	}
	code := string(bs)

	as, err := ioutil.ReadFile(abi + ".abi")
	if err != nil {
		return nil, err
	}

	var info contract.Info
	err = json.Unmarshal(as, &info)
	if err != nil {
		return nil, err
	}
	c := contract.Contract{
		ID:   id,
		Info: &info,
		Code: code,
	}

	return &c, nil
}

func array2json(ss []interface{}) string {
	x, err := json.Marshal(ss)
	if err != nil {
		panic(err)
	}
	return string(x)
}

var adminID string

func prepareContract(t *testing.T) *JSTester {
	js := NewJSTester(t)
	lc, err := ioutil.ReadFile("../contract/vote.js")
	if err != nil {
		t.Fatal(err)
	}
	bh := &block.BlockHead{
		ParentHash: []byte("abc"),
		Number:     0,
		Witness:    "witness",
		Time:       123456,
	}
	js.NewBlock(bh)

	js.SetJS(string(lc))
	js.SetAPI("RegisterProducer", "string", "string", "string", "string")
	js.SetAPI("UpdateProducer", "string", "string", "string", "string")
	js.SetAPI("LogInProducer", "string")
	js.SetAPI("LogOutProducer", "string")
	js.SetAPI("UnregisterProducer", "string")
	js.SetAPI("Vote", "string", "string", "number")
	js.SetAPI("Unvote", "string", "string", "number")
	js.SetAPI("InitProducer", "string")
	js.SetAPI("Stat")
	js.SetAPI("Init")
	for i := 0; i <= 18; i += 2 {
		js.vi.SetBalance(testID[i], 5e+7*1e8)
	}
	adminID = "IOSTrGdaqXePYMyo33DhjHthVSzFCmv7khwXejvBTcRvVbFoNjbrV"
	js.vi.SetBalance(adminID, 5e+7*1e8)
	js.vi.Commit()
	r := js.DoSet()
	if r.Status.Code != 0 {
		t.Fatal(r.Status.Message)
	}

	// deploy iost.bonus
	act2 := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.bonus", native.BonusABI().B64Encode()))
	trx2, err := MakeTx(act2)
	if err != nil {
		t.Fatal(err)
	}
	r, err = js.e.Exec(trx2, time.Second)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	// deploy iost.auth
	ca, err := Compile("iost.auth", "../contract/account", "../contract/account.js")
	act2 = tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.auth", ca.B64Encode()))
	trx2, err = MakeTx(act2)
	if err != nil {
		t.Fatal(err)
	}
	r, err = js.e.Exec(trx2, time.Second)
	if err != nil || r.Status.Code != tx.Success {
		t.Fatal(err, r)
	}

	for i := 0; i <= 18; i += 2 {
		ac, err := account.NewKeyPair(common.Base58Decode(testID[i+1]), crypto.Secp256k1)
		act2 := tx.NewAction("iost.auth", "SignUp", array2json([]interface{}{testID[i], testID[i], testID[i]}))
		trx2, err := MakeTxWithAuth(act2, ac)
		if err != nil {
			t.Fatal(err)
		}
		r, err = js.e.Exec(trx2, time.Second)
		if err != nil || r.Status.Code != tx.Success {
			t.Fatal(err, r)
		}

	}
	return js
}

func TestJS1_Vote1(t *testing.T) {
	t.Skip()
	ilog.Stop()
	js := prepareContract(t)
	defer js.Clear()

	for i := 6; i <= 18; i += 2 {
		if !(int64(40000000*1e8) < js.vi.Balance(testID[i])) {
			t.Fatal("error in balance :", i, js.vi.Balance(testID[i]))
		}
	}

	// test register, should success
	r := js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[0]))
	if r.Status.Code != 0 {
		t.Fatal(r.Status.Message)
	}

	// test require auth
	r = js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[2]))
	if r.Status.Code != 4 {
		t.Fatal(r.Status.Message)
	}

	// get pending producer info
	rtn := database.MustUnmarshal(js.vi.Get(js.cname + "-" + "pendingBlockNumber"))
	if rtn != "0" {
		t.Fatal(rtn)
	}
	srtn := js.ReadMap("producerTable", testID[0])
	if srtn != `{"loc":"loc","url":"url","netId":"netid","online":false,"score":0,"votes":0}` {
		t.Fatal(srtn)
	}
	// test re register
	r = js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[0]))
	if r.Status.Code != 4 {
		t.Fatal(r.Status.Message)
	}
}

//nolint
func TestJS_Vote(t *testing.T) {
	t.Skip()
	Convey("test of vote", t, func() {
		ilog.Stop()

		js := prepareContract(t)
		defer js.Clear()

		var r *tx.TxReceipt

		for i := 0; i < 14; i += 2 {
			tt := watchTime(func() {
				r = js.TestJS("InitProducer", fmt.Sprintf(`["%v"]`, testID[i]))
			})
			if r.Status.Code != 0 {
				t.Log(tt)
				t.Fatal(r.Status.Message)
			}
			t.Log(r.GasUsage)
			t.Log(tt)
		}

		keys := []string{
			"producerRegisterFee", "producerNumber", "preProducerThreshold", "preProducerMap",
			"voteLockTime", "currentProducerList", "pendingProducerList", "pendingBlockNumber",
			"producerTable",
			"voteTable",
		}
		_ = keys

		bh := &block.BlockHead{
			ParentHash: []byte("abc"),
			Number:     10,
			Witness:    "witness",
			Time:       123456,
		}
		js.NewBlock(bh)

		// test register, login, logout
		So(js.ReadMap("producerTable", testID[0]).(string), ShouldEqual, `{"loc":"","url":"","netId":"","online":true,"score":0,"votes":0}`)

		Convey("test of vote", func() {
			r := js.TestJS("LogOutProducer", `["a"]`)
			So(r.Status.Message, ShouldContainSubstring, "require auth failed")

			t.Log("time of log in", watchTime(func() {
				r = js.TestJSWithAuth("LogInProducer", fmt.Sprintf(`["%v"]`, testID[14]), testID[15])
			}))
			So(r.Status.Message, ShouldContainSubstring, "producer not exists")

			t.Log("time of register", watchTime(func() {
				r = js.TestJSWithAuth("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[14]), testID[15])
			}))
			So(r.Status.Message, ShouldEqual, "")

			r = js.TestJSWithAuth("LogInProducer", fmt.Sprintf(`["%v"]`, testID[14]), testID[15])
			So(r.Status.Message, ShouldEqual, "")

			r = js.TestJSWithAuth("LogOutProducer", fmt.Sprintf(`["%v"]`, testID[14]), testID[15])
			So(r.Status.Message, ShouldEqual, "")

			r = js.TestJS("UpdateProducer", fmt.Sprintf(`["%v", "%v", "%v", "%v"]`, testID[0], "nloc", "nurl", "nnetid"))
			So(r.Status.Message, ShouldEqual, "")

			So(js.ReadMap("producerTable", testID[0]).(string), ShouldEqual, `{"loc":"nloc","url":"nurl","netId":"nnetid","online":true,"score":0,"votes":0}`)

			// stat, no changes
			r = js.TestJS("Stat", `[]`)
			So(r.Status.Message, ShouldContainSubstring, "block number mismatch")

			So(js.ReadDB(`pendingProducerList`), ShouldEqual, `["IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn",`+
				`"IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd","IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY",`+
				`"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9","IOST7GmPn8xC1RESMRS6a62RmBcCdwKbKvk2ZpxZpcXdUPoJdapnnh",`+
				`"IOST7ZGQL4k85v4wAxWngmow7JcX4QFQ4mtLNjgvRrEnEuCkGSBEHN","IOST7ZNDWeh8pHytAZdpgvp7vMpjZSSe5mUUKxDm6AXPsbdgDMAYhs"]`)

			// vote and unvote
			r = js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 10000000))
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadMap("producerTable", testID[0]).(string), ShouldEqual, `{"loc":"nloc","url":"nurl","netId":"nnetid","online":true,"score":0,"votes":10000000}`)

			r = js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 10000000))
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadMap("producerTable", testID[0]).(string), ShouldEqual, `{"loc":"nloc","url":"nurl","netId":"nnetid","online":true,"score":0,"votes":20000000}`)

			r = js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[2], 10000000))
			So(r.Status.Message, ShouldContainSubstring, "require auth failed")

			r = js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 10000000))
			So(r.Status.Message, ShouldContainSubstring, "vote still locked")

			// stat
			r = js.TestJS("Stat", `[]`)
			So(r.Status.Message, ShouldContainSubstring, "block number mismatch")

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     200,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)
			t.Log("time of stat", watchTime(func() {
				r = js.TestJS("Stat", `[]`)
			}))
			if r.Status.Code != 0 {
				t.Fatal(r.Status.Message)
			}

			// 0, 6, 2, 12, 8, 10, 4
			So(js.ReadDB(`pendingProducerList`), ShouldEqual, `["IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn",`+
				`"IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd","IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY",`+
				`"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9","IOST7GmPn8xC1RESMRS6a62RmBcCdwKbKvk2ZpxZpcXdUPoJdapnnh",`+
				`"IOST7ZGQL4k85v4wAxWngmow7JcX4QFQ4mtLNjgvRrEnEuCkGSBEHN","IOST7ZNDWeh8pHytAZdpgvp7vMpjZSSe5mUUKxDm6AXPsbdgDMAYhs"]`)

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     211,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)

			// test unvote
			r = js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 20000001))
			So(r.Status.Message, ShouldContainSubstring, "vote amount less than expected")

			r = js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 1000000))
			So(r.Status.Message, ShouldEqual, "")

			So(js.vi.Servi(testID[0]), ShouldEqual, int64(1055000*1e8))
			So(js.vi.TotalServi(), ShouldEqual, int64(1055000*1e8))
			// stat pending producers don't get score

			// seven
			for i := 16; i <= 18; i += 2 {
				r = js.TestJSWithAuth("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[i]), testID[i+1])
				So(r.Status.Message, ShouldEqual, "")
			}

			for i := 2; i <= 18; i += 2 {
				r = js.TestJSWithAuth("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[i], testID[i], 30000000+i), testID[i+1])
				So(r.Status.Message, ShouldEqual, "")
				So(js.ReadMap("producerTable", testID[i]), ShouldContainSubstring, strconv.Itoa(30000000+i))
			}
			for i := 14; i <= 18; i += 2 {
				So(js.ReadMap("preProducerMap", testID[i]), ShouldEqual, "true")
			}

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     400,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)

			// stat, offline producers and pending producers don't get score
			r = js.TestJS("Stat", `[]`)
			So(r.Status.Message, ShouldEqual, "")

			for i := 2; i <= 18; i += 2 {
				So(js.ReadMap("producerTable", testID[i]), ShouldContainSubstring, `"score":0`)
			}

			for i := 14; i <= 18; i += 2 {
				r = js.TestJSWithAuth("LogInProducer", fmt.Sprintf(`["%v"]`, testID[i]), testID[i+1])
				So(r.Status.Message, ShouldEqual, "")
			}

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     600,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)

			// stat, 1 producer become pending
			t.Log("time of stat", watchTime(func() {
				r = js.TestJS("Stat", `[]`)
				// 14, 16, 18 get score, 18 in, 4 out
			}))
			So(r.Status.Message, ShouldEqual, "")

			for i := 14; i <= 18; i += 2 {
				So(js.ReadMap("producerTable", testID[i]), ShouldContainSubstring, fmt.Sprintf(`"score":%d`, 9000000+i))
			}

			So(js.ReadDB("pendingProducerList"), ShouldContainSubstring, testID[18])

			// stat, offline producer doesn't become pending.
			r = js.TestJSWithAuth("LogOutProducer", fmt.Sprintf(`["%v"]`, testID[16]), testID[17])
			So(r.Status.Message, ShouldEqual, "")

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     800,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)

			r = js.TestJS("Stat", `[]`)
			// 4, 14 get score, 14 in, 10 out
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadDB(`pendingProducerList`), ShouldEqual, `["IOST8mFxe4kq9XciDtURFZJ8E76B8UssBgRVFA5gZN9HF5kLUVZ1BB",`+
				`"IOST6wYBsLZmzJv22FmHAYBBsTzmV1p1mtHQwkTK9AjCH9Tg5Le4i4","IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn",`+
				`"IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd","IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY",`+
				`"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9","IOST7GmPn8xC1RESMRS6a62RmBcCdwKbKvk2ZpxZpcXdUPoJdapnnh"]`)

			r = js.TestJSWithAuth("LogInProducer", fmt.Sprintf(`["%v"]`, testID[16]), testID[17])
			So(r.Status.Message, ShouldEqual, "")

			// stat, offline producer doesn't become pending. offline and pending producer don't get score, other pre producers get score
			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     1000,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)
			r = js.TestJS("Stat", `[]`)
			// 4, 10, 16 get score, 16 in, 8 out
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadDB("pendingProducerList"), ShouldContainSubstring, testID[16])

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     1200,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)
			r = js.TestJS("Stat", `[]`)
			// 4, 8, 10 get score, 4 in, 12 out
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadDB("pendingProducerList"), ShouldContainSubstring, testID[4])

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     1400,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)
			r = js.TestJS("Stat", `[]`)
			// 8, 10, 12 get score, 10 in, 2 out
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadDB("pendingProducerList"), ShouldContainSubstring, testID[10])

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     1600,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)
			r = js.TestJS("Stat", `[]`)
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadDB("pendingProducerList"), ShouldContainSubstring, testID[8])
			So(js.ReadDB("pendingProducerList"), ShouldNotContainSubstring, testID[6])

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     1800,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)
			r = js.TestJS("Stat", `[]`)
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadDB("pendingProducerList"), ShouldContainSubstring, testID[12])

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     2000,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)
			r = js.TestJS("Stat", `[]`)
			So(r.Status.Message, ShouldEqual, "")
			So(js.ReadDB("pendingProducerList"), ShouldContainSubstring, testID[4])
			So(js.ReadDB("pendingProducerList"), ShouldNotContainSubstring, testID[18])

			// unregister
			r = js.TestJS("UnregisterProducer", fmt.Sprintf(`["%v"]`, testID[0]))
			So(r.Status.Message, ShouldEqual, "")

			// unvote after unregister
			r = js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 9000000))
			So(r.Status.Message, ShouldEqual, "")

			So(js.vi.Servi(testID[0]), ShouldEqual, 91055000*1e8)
			So(js.vi.TotalServi(), ShouldEqual, 91055000*1e8)

			// re register, score = 0, vote = 0
			r = js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[0]))
			So(r.Status.Message, ShouldEqual, "")
			r = js.TestJS("LogInProducer", fmt.Sprintf(`["%v"]`, testID[0]))
			So(r.Status.Message, ShouldEqual, "")

			js.vi.SetBalance(testID[2], 5e+7*1e8)
			r = js.TestJSWithAuth("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[2], 21000001), testID[3])
			So(r.Status.Message, ShouldEqual, "")

			bh = &block.BlockHead{
				ParentHash: []byte("abc"),
				Number:     2200,
				Witness:    "witness",
				Time:       123456,
			}
			js.NewBlock(bh)
			r = js.TestJS("Stat", `[]`)
			So(r.Status.Message, ShouldEqual, "")

			// unregister pre producer
			r = js.TestJSWithAuth("UnregisterProducer", fmt.Sprintf(`["%v"]`, testID[10]), testID[11])
			So(r.Status.Message, ShouldContainSubstring, "can't unregist")

			//// test bonus
			//act2 := tx.NewAction("iost.bonus", "ClaimBonus", fmt.Sprintf(`["%v", %d]`, testID[0], 1))
			//trx2, err := MakeTx(act2)
			//if err != nil {
			//	t.Fatal(err)
			//}
			//r, err = js.e.Exec(trx2, time.Second)
			//if err != nil || r.Status.Code != tx.Success {
			//	t.Fatal(err, r)
			//}
			//
			//So(js.vi.Servi(testID[0]), ShouldEqual, 91054999)
			//So(js.vi.Balance(testID[0]), ShouldEqual, 3900000000876051)
			//So(js.vi.Balance(host.ContractAccountPrefix+"iost.bonus"), ShouldEqual, 23299)
			//act2 = tx.NewAction("iost.bonus", "ClaimBonus", fmt.Sprintf(`["%v", %d]`, testID[0], 91054999))
			//
			//trx2, err = MakeTx(act2)
			//if err != nil {
			//	t.Fatal(err)
			//}
			//r, err = js.e.Exec(trx2, time.Second)
			//if err != nil || r.Status.Code != tx.Success {
			//	t.Fatal(err, r)
			//}
			//
			//So(js.vi.Servi(testID[0]), ShouldEqual, 0)
			//So(js.vi.Balance(host.ContractAccountPrefix+"iost.bonus"), ShouldEqual, 620)
			//So(js.vi.Balance(testID[0]), ShouldEqual, 3900000000492070)
		})

		Convey("test of vote update", func() {
			t.Skip()

			fd, err := ioutil.ReadFile("./test_data/vote_update.js")
			if err != nil {
				t.Fatal(err)
			}
			rawCode := string(fd)
			fd, err = ioutil.ReadFile("./test_data/vote_update.js.abi")
			if err != nil {
				t.Fatal(err)
			}
			rawAbi := string(fd)
			c := contract.Compiler{}
			code, err := c.Parse(js.cname, rawCode, rawAbi)
			if err != nil {
				t.Fatal(err)
			}

			act := tx.NewAction("iost.system", "UpdateCode", fmt.Sprintf(`["%v", "%v"]`, code.B64Encode(), ""))

			trx := tx.NewTx([]*tx.Action{act}, nil, 100000, 100, 10000000, 0)

			ac, err := account.NewKeyPair(common.Base58Decode("37qTTtYLMt7FirFxVxYGDD547hZtRw7MpAyeoiJRF72hVXiWwBCz3AzCxeFnPuHaULxz3jT8sQg93EofBBBr99Q9"), crypto.Ed25519)
			So(account.GetIDByPubkey(ac.Pubkey), ShouldEqual, adminID)
			if err != nil {
				t.Fatal(err)
			}
			trx, err = tx.SignTx(trx, ac.ID, ac)
			if err != nil {
				t.Fatal(err)
			}

			r, err := js.e.Exec(trx, time.Second)
			if err != nil || r.Status.Code != tx.Success {
				t.Fatal(err, r)
			}

			r, err = js.e.Exec(trx, time.Second)
			So(err, ShouldEqual, nil)
			So(r.Status.Code, ShouldEqual, tx.ErrorRuntime)
			So(r.Status.Message, ShouldContainSubstring, "update refused")

		})
	})
}

//nolint
func TestJS_Genesis(t *testing.T) {
	t.Skip("skip genesis")

	witnessInfo := testID
	var acts []*tx.Action
	for i := 0; i < len(witnessInfo)/2; i++ {
		act := tx.NewAction("iost.system", "IssueIOST", fmt.Sprintf(`["%v", %v]`, witnessInfo[2*i], 50000000))
		acts = append(acts, act)
	}
	VoteContractPath := os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/config/"
	// deploy iost.vote
	voteFilePath := VoteContractPath + "vote.js"
	voteAbiPath := VoteContractPath + "vote.js.abi"
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
	num := len(witnessInfo) / 2

	act := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.vote", code.B64Encode()))
	acts = append(acts, act)

	for i := 0; i < num; i++ {
		act1 := tx.NewAction("iost.vote", "InitProducer", fmt.Sprintf(`["%v"]`, witnessInfo[2*i]))
		acts = append(acts, act1)
	}

	// deploy iost.bonus
	act2 := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["%v", "%v"]`, "iost.bonus", native.BonusABI().B64Encode()))
	acts = append(acts, act2)

	trx := tx.NewTx(acts, nil, 10000000, 0, 0, 0)
	trx.Time = 0
	acc, err := account.NewKeyPair(common.Base58Decode("BQd9x7rQk9Y3rVWRrvRxk7DReUJWzX4WeP9H9H4CV8Mt"), crypto.Secp256k1)
	if err != nil {
		t.Fatal(err)
	}
	trx, err = tx.SignTx(trx, acc.ID, acc)
	if err != nil {
		t.Fatal(err)
	}

	blockHead := block.BlockHead{
		Version:    0,
		ParentHash: nil,
		Number:     0,
		Witness:    acc.ID,
		Time:       time.Now().Unix() / common.SlotLength,
	}
	mvccdb, err := db.NewMVCCDB("mvcc")
	defer closeMVCCDB(mvccdb)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(&blockHead, mvccdb)
	engine.SetUp("js_path", os.Getenv("GOPATH")+"/src/github.com/iost-official/go-iost/vm/v8vm/v8/libjs/")
	var txr *tx.TxReceipt
	ti := watchTime(func() {
		txr, err = engine.Exec(trx, time.Second)
	})
	if err != nil {
		t.Fatal(fmt.Errorf("exec tx failed, stop the pogram. err: %v", err))
	}
	if txr.Status.Code != 0 {
		t.Fatal(txr.Status.Message)
	}
	if ti > time.Second {
		t.Fatal(ti)
	}
	//pl := database.MustUnmarshal(database.NewVisitor(0, mvccdb).Get("iost.vote" + "-" + "pendingProducerList"))

	if txr.Status.Code != tx.Success {
		t.Fatal("exec trx failed.")
	}
	blk := block.Block{
		Head:     &blockHead,
		Sign:     &crypto.Signature{},
		Txs:      []*tx.Tx{trx},
		Receipts: []*tx.TxReceipt{txr},
	}
	blk.Head.TxsHash = blk.CalculateTxsHash()
	blk.Head.MerkleHash = blk.CalculateMerkleHash()
	err = blk.CalculateHeadHash()
	if err != nil {
		t.Fatal(err)
	}
}
