package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/vm/v8vm"
)

var vmPool *v8.VMPool

func init() {
	vmPool = v8.NewVMPool(3, 100)
	vmPool.SetJSPath("../v8/libjs/")
	vmPool.Init()
}

// MyInit init host and contract
func MyInit(conName string, optional ...interface{}) (*host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath("simple.json")
	vi := database.NewVisitor(100, db)

	ctx := host.NewContext(nil)
	ctx.Set("gas_ratio", int64(100))
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

	var times = 10000
	if len(os.Args) >= 2 {
		timesT, err := strconv.Atoi(os.Args[1])
		times = timesT
		if err != nil {
		}
	}

	fmt.Println("runnig now...")
	a := time.Now()

	var i = 0
	for ; i < times; i++ {
		expTime := time.Now().Add(time.Second * 10)
		host.SetDeadline(expTime)
		_, _, err := vmPool.LoadAndCall(host, code, "show")
		if err != nil {
			log.Fatal("error: ", err)
		}
	}
	tps := float64(times) / time.Since(a).Seconds()
	fmt.Println("time used: ", time.Since(a))
	fmt.Println("each: ", tps)
}
