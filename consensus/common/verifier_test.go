package consensus_common

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
)

func TestStdTxsVerifier(t *testing.T) {
	dbx, err := db.DatabaseFactor("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	main := lua.NewMethod(2, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`

	txs := make([]tx.Tx, 0)
	for j := 0; j < 100; j++ {
		lc := lua.NewContract(vm.ContractInfo{Prefix: strconv.Itoa(j), GasLimit: 100, Price: 0, Publisher: vm.IOSTAccount("a")}, code, main)
		txx := tx.NewTx(int64(j), &lc)
		txs = append(txs, txx)
	}

	fmt.Println("len", len(txs))
	//fmt.Println(txs[0].Contract)
	p2, _, err := StdTxsVerifier(txs, pool)
	//fmt.Println(p2.GetPatch())
	//fmt.Println(p2.Get("mark"))
	p := p2.(*state.PoolImpl)
	count := 0
	for p != nil {
		p = p.Parent()
		count++
	}
	fmt.Println("depth:", count)
}

func BenchmarkStdTxsVerifier(b *testing.B) {
	dbx, err := db.DatabaseFactor("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	main := lua.NewMethod(2, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`

	txs := make([]tx.Tx, 0)
	for j := 0; j < 10000; j++ {
		lc := lua.NewContract(vm.ContractInfo{Prefix: strconv.Itoa(j), GasLimit: 100, Price: 0, Publisher: vm.IOSTAccount("a")}, code, main)
		txx := tx.NewTx(int64(j), &lc)
		txs = append(txs, txx)
	}

	//fmt.Println(txs[500].Contract)
	var k int
	for i := 0; i < b.N; i++ {
		pool, k, _ = StdTxsVerifier(txs, pool)
	}
	fmt.Println(pool.GetPatch())

	fmt.Println(k)

}
