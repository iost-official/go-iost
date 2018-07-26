package main

import (
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/verifier"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	"github.com/iost-official/Go-IOS-Protocol/vm/lua"
)

func main() {
	main := lua.NewMethod(vm.Public, "main", 0, 1)
	code := `function main()
	Transfer("a", "b", 50)
end`
	lc := lua.NewContract(vm.ContractInfo{Prefix: "test", GasLimit: 100, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)

	dbx, err := db.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	pool.PutHM(state.Key("iost"), state.Key("a"), state.MakeVFloat(1000000))
	pool.PutHM(state.Key("iost"), state.Key("b"), state.MakeVFloat(1000000))
	fmt.Println("--------------------")
	//fmt.Println(pool.GetHM("iost", "b"))
	var pool2 state.Pool

	cv := verifier.NewCacheVerifier()
	pool2, err = cv.VerifyContract(&lc, pool)
	if err != nil {
		panic(err)
	}
	aa, err := pool2.GetHM("iost", "a")
	ba, err := pool2.GetHM("iost", "b")
	fmt.Println(aa)
	fmt.Println(ba)
}
