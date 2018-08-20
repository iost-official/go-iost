package new_vm

import (
	"testing"

	"os"

	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

func benchInit() (Engine, *database.Visitor) {
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
	trx, err := makeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
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
	trx, err := makeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx)
	}
	b.StopTimer()
	cleanUp()
}

func BenchmarkNative_Receipt(b *testing.B) {
	e, _ := benchInit()

	act := tx.NewAction("iost.system", "Receipt", `["my receipt"]`)
	trx, err := makeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
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
	trx, err := makeTx(act)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		e.Exec(trx)
	}
	b.StopTimer()
	cleanUp()
}
