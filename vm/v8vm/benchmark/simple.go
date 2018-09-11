package main

import (
	"fmt"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
	"github.com/iost-official/Go-IOS-Protocol/vm/v8vm"
	"github.com/prometheus/common/log"
)

var vmPool *v8.VMPool

func init() {
	vmPool = v8.NewVMPool(3, 100)
	vmPool.SetJSPath("../v8/libjs/")
	vmPool.Init()
}

func MyInit(conName string, optional ...interface{}) (*host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath("simple.json")
	vi := database.NewVisitor(100, db)

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", int64(1))
	var gasLimit = int64(10000)
	if len(optional) > 0 {
		gasLimit = optional[0].(int64)
	}
	ctx.GSet("gas_limit", gasLimit)
	ctx.Set("contract_name", conName)
	h := host.NewHost(ctx, vi, nil, ilog.DefaultLogger())

	code := &contract.Contract{
		ID: conName,
		Code: `
class Contract {
	constructor() {
	}

	init() {
	}

	show() {
		return "hello world";
	}
}

module.exports = Contract;
`,
	}

	code.Code, _ = vmPool.Compile(code)

	return h, code
}

func main() {
	host, code := MyInit("simple")
	//vmPool.LoadAndCall(host, code, "show")

	var times float64 = 1000

	fmt.Println("runnig now...")
	a := time.Now()

	var i float64 = 0
	for ; i < times; i++ {
		_, _, err := vmPool.LoadAndCall(host, code, "show")
		if err != nil {
			log.Fatal("error: ", err)
		}
	}
	timeUsed := time.Since(a).Nanoseconds()
	tps := int(1000 / (float64(timeUsed) / 1000000 / times))
	fmt.Println("time used: ", time.Since(a))
	fmt.Println("each: ", tps)
}
