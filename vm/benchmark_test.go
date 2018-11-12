package vm

import (
	"testing"

	"os"

	"fmt"

	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"io/ioutil"
)

func benchInit() (Engine, *database.Visitor) {
	ilog.Stop()
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		panic(err)
	}

	vi := database.NewVisitor(0, mvccdb)
	vi.SetTokenBalance("iost", testID[0], 1000000000000)
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

func BenchmarkNative_Transfer(b *testing.B) { // 21400 ns/op
	e, _ := benchInit()

	act := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v","%v", 100]`, testID[0], testID[2]))
	trx, err := MakeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx, time.Second)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkNative_Transfer_LRU(b *testing.B) { // 15300 ns/op
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		panic(err)
	}

	vi := database.NewVisitor(100, mvccdb)
	vi.SetTokenBalance("iost", testID[0], 1000000000000)
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
		e.Exec(trx, time.Second)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkNative_Receipt(b *testing.B) { // 138000 ns/op
	e, _ := benchInit()

	act := tx.NewAction("iost.system", "Receipt", `["my receipt"]`)
	trx, err := MakeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx, time.Second)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkNative_SetCode(b *testing.B) { // 3.03 ms/op
	e, _ := benchInit()

	hw := jsHelloWorld()

	act := tx.NewAction("iost.system", "SetCode", fmt.Sprintf(`["%v"]`, hw.B64Encode()))
	trx, err := MakeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx, time.Second)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkJS_Gas_Once(b *testing.B) { // 443 us/op
	ilog.Stop()
	js := NewJSTester(b)
	defer js.Clear()
	f, err := ioutil.ReadFile("test_data/gas.js")
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
		r, err := js.e.Exec(trx2, time.Second)
		if r.Status.Code != 0 || err != nil {
			b.Fatal(r.Status.Message, err)
		}
	}
	b.StopTimer()
}

func BenchmarkJS_Gas_100(b *testing.B) { // 483 um/op
	ilog.Stop()
	js := NewJSTester(b)
	js.vi.SetTokenBalance("iost", testID[0], 10000000000)
	defer js.Clear()
	f, err := ioutil.ReadFile("test_data/gas.js")
	if err != nil {
		b.Fatal(err)
	}
	js.SetJS(string(f))
	js.SetAPI("run", "number")
	js.DoSet()

	act2 := tx.NewAction(js.cname, "run", `[100]`)

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
		r, err := js.e.Exec(trx2, time.Second)
		if r.Status.Code != 0 || err != nil {
			b.Fatal(r.Status.Message, err)
		}
	}
	b.StopTimer()
}

func BenchmarkJS_Gas_200(b *testing.B) { // 525 um/op
	ilog.Stop()
	js := NewJSTester(b)
	js.vi.SetTokenBalance("iost", testID[0], 10000000000)
	defer js.Clear()
	f, err := ioutil.ReadFile("test_data/gas.js")
	if err != nil {
		b.Fatal(err)
	}
	js.SetJS(string(f))
	js.SetAPI("run", "number")
	js.DoSet()

	act2 := tx.NewAction(js.cname, "run", `[200]`)

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
		r, err := js.e.Exec(trx2, time.Second)
		if r.Status.Code != 0 || err != nil {
			b.Fatal(r.Status.Message, err)
		}
	}
	b.StopTimer()
}

func Benchmark_JS_Transfer(b *testing.B) {
	ilog.Stop()
	js := NewJSTester(b)
	defer js.Clear()
	f, err := ioutil.ReadFile("test_data/transfer.js")
	if err != nil {
		b.Fatal(err)
	}
	js.SetJS(string(f))
	js.SetAPI("transfer", "string", "string", "number")
	js.DoSet()

	js.vi.SetTokenBalance("iost", testID[0], 100000000)

	act2 := tx.NewAction(js.cname, "transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], 100))

	ac, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		panic(err)
	}

	trx2, err := MakeTxWithAuth(act2, ac)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, err := js.e.Exec(trx2, time.Second)
		if r.Status.Code != 0 || err != nil {
			b.Fatal(r.Status.Message, err)
		}
	}
	b.StopTimer()
}
