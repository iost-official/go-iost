package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/core/version"
	"github.com/iost-official/go-iost/v3/ilog"
	"github.com/iost-official/go-iost/v3/vm/database"
	"github.com/iost-official/go-iost/v3/vm/host"
	v8 "github.com/iost-official/go-iost/v3/vm/v8vm"
)

var vmPool *v8.VMPool

func init() {
	vmPool = v8.NewVMPool(3, 100)
	vmPool.SetJSPath("../v8/libjs/")
	vmPool.Init()
}

// MyInit init host and contract
func MyInit(conName string, optional ...any) (*host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath("simple.json")
	vi := database.NewVisitor(100, db, version.NewRules(0))

	ctx := host.NewContext(nil)
	ctx.Set("gas_ratio", int64(100))
	var gasLimit = int64(10000)
	if len(optional) > 0 {
		gasLimit = optional[0].(int64)
	}
	ctx.GSet("gas_limit", gasLimit)
	ctx.Set("contract_name", conName)
	h := host.NewHost(ctx, vi, version.NewRules(0), nil, ilog.DefaultLogger())

	code := &contract.Contract{
		ID: conName,
		Code: `
class Contract {
	constructor() {
	}

	init() {
	}

	show() {
		return "Hello World!"
	}

	test() {
		let a = "what the fuck is this!"
		let b = new Array(10000)
		let d = 1;
		for (let i = 0; i < b.length; i++)
		{
			b[i] = new Array(1000)
		}
		return a
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
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var times = *flag.Int("times", 0, "write cpu profile to file") // nolint
	flag.Parse()
	// 如果命令行设置了 cpuprofile
	if *cpuprofile != "" {
		// 根据命令行指定文件名创建 profile 文件
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		// 开启 CPU profiling
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if times == 0 {
		times = 10000
	}

	fmt.Println("runnig now...")
	a := time.Now()

	var i = 0
	for ; i < times; i++ {
		expTime := time.Now().Add(time.Second * 10)
		host.SetDeadline(expTime)
		_, _, err := vmPool.LoadAndCall(host, code, "show")
		if err != nil {
			log.Fatal("error: ", err) // nolint
		}
	}
	tps := float64(times) / time.Since(a).Seconds()
	fmt.Println("time used: ", time.Since(a))
	fmt.Println("tps: ", tps)
}
