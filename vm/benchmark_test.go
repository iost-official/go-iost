package vm

import (
	"testing"

	"os"

	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

func benchInit() (Engine, *database.Visitor) {
	ilog.Stop()
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		panic(err)
	}

	vi := database.NewVisitor(0, mvccdb)
	vi.SetBalance(testID[0], 1000000)
	vi.SetContract(systemContract)
	vi.Commit()

	bh := &block.BlockHead{
		ParentHash: []byte("abc"),
		Number:     10,
		Witness:    "witness",
		Time:       123456,
	}

	e := newEngine(bh, vi)

	e.SetUp("js_path", jsPath)
	e.SetUp("log_level", "fatal")
	e.SetUp("log_enable", "")
	return e, vi
}

func cleanUp() {
	os.RemoveAll("mvcc")
}

func BenchmarkNative_Transfer(b *testing.B) {
	e, _ := benchInit()

	act := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v","%v", 100]`, testID[0], testID[2]))
	trx, err := MakeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkNative_Transfer_LRU(b *testing.B) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		panic(err)
	}

	vi := database.NewVisitor(100, mvccdb)
	vi.SetBalance(testID[0], 1000000)
	vi.SetContract(systemContract)
	vi.Commit()

	bh := &block.BlockHead{
		ParentHash: []byte("abc"),
		Number:     10,
		Witness:    "witness",
		Time:       123456,
	}

	e := newEngine(bh, vi)

	e.SetUp("js_path", jsPath)
	e.SetUp("log_level", "fatal")
	e.SetUp("log_enable", "")

	act := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v","%v", 100]`, testID[0], testID[2]))
	trx, err := MakeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkNative_Receipt(b *testing.B) {
	e, _ := benchInit()

	act := tx.NewAction("iost.system", "Receipt", `["my receipt"]`)
	trx, err := MakeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkNative_SetCode(b *testing.B) {
	e, _ := benchInit()

	hw := jsHelloWorld()

	act := tx.NewAction("iost.system", "SetCode", fmt.Sprintf(`["%v"]`, hw.B64Encode()))
	trx, err := MakeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkJS_Gas_Once(b *testing.B) {
	ilog.Stop()
	js := NewJSTester(b)
	defer js.Clear()
	f, err := ReadFile("test_data/gas.js")
	if err != nil {
		b.Fatal(err)
	}
	js.SetJS(string(f))
	js.SetAPI("run", "number")
	js.DoSet()

	act2 := tx.NewAction(js.cname, "run", `[1]`)

	trx2, err := MakeTx(act2)
	if err != nil {
		js.t.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// r := js.TestJS("single", `[]`)
		//if i == 0 {
		//	b.Log("gas is : ", r.GasUsage)
		//}
		js.e.Exec(trx2)
	}
	b.StopTimer()
}

func Benchmark_JS_Transfer(b *testing.B) {
	ilog.Stop()
	js := NewJSTester(b)
	defer js.Clear()
	f, err := ReadFile("test_data/transfer.js")
	if err != nil {
		b.Fatal(err)
	}
	js.SetJS(string(f))
	js.SetAPI("transfer", "string", "string", "number")
	js.DoSet()

	js.vi.SetBalance(testID[0], 100000000)

	act2 := tx.NewAction(js.cname, "transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], 100))

	ac, err := account.NewAccount(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		panic(err)
	}

	trx2, err := MakeTxWithAuth(act2, ac)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, err := js.e.Exec(trx2)
		if r.Status.Code != 0 || err != nil {
			b.Fatal(r.Status.Message, err)
		}
	}
	b.StopTimer()
}

func BenchmarkVote_InitProducer(b *testing.B) {
	ilog.Stop()
	js := NewJSTester(b)
	bh := block.BlockHead{
		Number: 0,
	}
	js.NewBlock(&bh)
	js.vi.SetBalance(testID[0], 100000000)
	defer js.Clear()
	f, err := ReadFile("../config/vote.js")
	if err != nil {
		b.Fatal(err)
	}
	js.SetJS(string(f))
	js.SetAPI("InitProducer", "string")
	js.DoSet()

	js.vi.SetBalance(testID[0], 100000000)

	act1 := tx.NewAction(js.cname, "InitProducer", fmt.Sprintf(`["%v"]`, testID[0]))

	ac, err := account.NewAccount(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		panic(err)
	}

	trx2, err := MakeTxWithAuth(act1, ac)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, err := js.e.Exec(trx2)
		if r.Status.Code != 0 || err != nil {
			b.Fatal(r.Status.Message, err)
		}
	}
	b.StopTimer()
}
