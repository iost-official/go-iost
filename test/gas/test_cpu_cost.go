package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	v8 "github.com/iost-official/go-iost/vm/v8vm"
	"github.com/wcharczuk/go-chart"
)

var (
	// OpList is the list of all operation
	OpList = map[string][]string{
		"empty": {
			"Empty",
		},
		"base": {
			"Call",
			"New",
			"Throw",
			"Yield",
			"Member",
			"Meta",
			"Assignment",
			"Plus",
			"Add",
			"Sub",
			"Mutiple",
			"Div",
			"Not",
			"And",
			"Conditional",
		},
		"lib": {
			"StringCharAt",
			"StringCharCodeAt",
			"StringLength",
			"StringConstructor",
			"StringToString",
			"StringValueOf",
			"StringConcat",
			"StringIncludes",
			"StringEndsWith",
			"StringIndexOf",
			"StringLastIndexOf",
			"StringReplace",
			"StringSearch",
			"StringSplit",
			"StringStartsWith",
			"StringSlice",
			"StringSubstring",
			"StringToLowerCase",
			"StringToUpperCase",
			"StringTrim",
			"StringTrimLeft",
			"StringTrimRight",
			"StringRepeat",
			"ArrayIsArray",
			"ArrayOf",
			"ArrayConcat",
			"ArrayEvery",
			"ArrayFilter",
			"ArrayFind",
			"ArrayFindIndex",
			"ArrayForEach",
			"ArrayIncludes",
			"ArrayIndexOf",
			"ArrayJoin",
			"ArrayKeys",
			"ArrayLastIndexOf",
			"ArrayMap",
			"ArrayPop",
			"ArrayPush",
			"ArrayReverse",
			"ArrayShift",
			"ArraySlice",
			"ArraySort",
			"ArraySplice",
			"ArrayToString",
			"ArrayUnshift",
			"JSONParse",
			"JSONStringify",
		},
		"storage": {
			"Put",
			"Get",
			"Has",
			"Del",
			"MapPut",
			"MapGet",
			"MapHas",
			"MapDel",
			"MapKeys",
			"MapLen",
		},
	}
)

var vmPool *v8.VMPool
var testDataPath = "./test_data/"
var baseCPUCost = int64(2000)

func runOp(vi *database.Visitor, name string, api string, num int) (float64, int64) {
	b, err := ioutil.ReadFile(path.Join(testDataPath, name))
	if err != nil {
		log.Fatalf("Read file failed: %v", err)
	}
	code := string(b)

	now := time.Now()

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", int64(1))
	ctx.GSet("gas_limit", int64(100000000))
	ctx.Set("contract_name", name)

	host := host.NewHost(ctx, vi, nil, ilog.DefaultLogger())
	expTime := time.Now().Add(time.Second * 10)
	host.SetDeadline(expTime)

	contract := &contract.Contract{
		ID:   name,
		Code: code,
	}

	contract.Code, err = vmPool.Compile(contract)
	if err != nil {
		log.Fatalf("Compile contract failed: %v", err)
	}

	_, cost, err := vmPool.LoadAndCall(host, contract, api, num)

	if err != nil {
		log.Fatalf("LoadAndCall %v.%v %v failed: %v", contract, api, num, err)
	}

	return time.Now().Sub(now).Seconds(), baseCPUCost + cost.CPU
}

func init() {
	// TODO The number of pool need adjust
	vmPool = v8.NewVMPool(30, 30)
	vmPool.Init()
}

func getOpDetail() {
	mvccdb, err := db.NewMVCCDB("mvccdb")
	if err != nil {
		log.Fatalf("New MVCC DB failed: %v", err)
	}
	vi := database.NewVisitor(100, mvccdb)

	for _, opType := range []string{"empty", "base", "lib", "storage"} {
		for _, op := range OpList[opType] {
			fmt.Printf("Start %v:%v...\n", opType, op)
			x := make([]float64, 0)
			yt := make([]float64, 0)
			yc := make([]float64, 0)
			for i := 0; ; i = i + 10000 {
				tcost, ccost := runOp(
					vi,
					fmt.Sprintf("%v_op.js", opType),
					fmt.Sprintf("do%v", op),
					i,
				)

				x = append(x, float64(i))
				yt = append(yt, tcost*1000)
				yc = append(yc, float64(ccost))
				if tcost > 0.2 {
					break
				}
			}

			graph := chart.Chart{
				XAxis: chart.XAxis{
					Style: chart.StyleShow(),
				},
				YAxis: chart.YAxis{
					Style: chart.StyleShow(),
					Range: &chart.ContinuousRange{
						Min: 0.0,
						Max: 200.0,
					},
				},
				YAxisSecondary: chart.YAxis{
					Style: chart.StyleShow(),
					Range: &chart.ContinuousRange{
						Min: 0.0,
						Max: 20000000.0,
					},
				},
				Series: []chart.Series{
					chart.ContinuousSeries{
						XValues: x,
						YValues: yt,
					},
					chart.ContinuousSeries{
						YAxis:   chart.YAxisSecondary,
						XValues: x,
						YValues: yc,
					},
				},
			}

			f, err := os.Create(fmt.Sprintf("%s/%s.png", opType, op))
			if err != nil {
				log.Fatal(err)
			}
			if err := graph.Render(chart.PNG, f); err != nil {
				log.Fatal(err)
			}
			//fmt.Printf("Time: %0.3fs\n", tcost)
			//fmt.Printf("CPU Cost: %vgas\n", ccost)
		}
	}

	os.RemoveAll("mvccdb")
}

// OpInfo is the information of operation for overview test
type OpInfo struct {
	Name string
	Time float64
	Gas  float64
}

func getOverview() {
	mvccdb, err := db.NewMVCCDB("mvccdb")
	if err != nil {
		log.Fatalf("New MVCC DB failed: %v", err)
	}
	vi := database.NewVisitor(100, mvccdb)

	OpInfos := make([]*OpInfo, 0)

	for _, opType := range []string{"base", "lib", "storage"} {
		for _, op := range OpList[opType] {
			fmt.Printf("Start %v:%v...\n", opType, op)
			for i := 0; ; i = i + 10000 {
				tcost, ccost := runOp(
					vi,
					fmt.Sprintf("%v_op.js", opType),
					fmt.Sprintf("do%v", op),
					i,
				)

				if tcost > 0.2 {
					emptyT, emptyC := runOp(
						vi,
						fmt.Sprintf("%v_op.js", "empty"),
						fmt.Sprintf("do%v", "Empty"),
						i,
					)
					OpInfos = append(OpInfos,
						&OpInfo{
							Name: fmt.Sprintf("%v:%v", opType, op),
							Time: (tcost - emptyT) * 1e9 / float64(i),
							Gas:  float64(ccost-emptyC) / float64(i),
						},
					)
					break
				}
			}
		}
	}

	barsT := make([]chart.Value, 0)
	barsG := make([]chart.Value, 0)
	barsC := make([]chart.Value, 0)
	for _, opInfo := range OpInfos {
		barsT = append(barsT,
			chart.Value{
				Value: opInfo.Time,
				Label: opInfo.Name,
			},
		)
		barsG = append(barsG,
			chart.Value{
				Value: opInfo.Gas,
				Label: opInfo.Name,
			},
		)
		barsC = append(barsC,
			chart.Value{
				Value: opInfo.Gas / opInfo.Time * 1000,
				Label: opInfo.Name,
			},
		)
	}

	graph := chart.BarChart{
		Title:      "Time (ns)",
		TitleStyle: chart.StyleShow(),
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:     2048,
		Width:      8192,
		BarWidth:   100,
		BarSpacing: 10,
		XAxis: chart.Style{
			Show:     true,
			FontSize: 6,
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		Bars: barsT,
	}

	ft, err := os.Create("overview_time.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := graph.Render(chart.PNG, ft); err != nil {
		log.Fatal(err)
	}

	fg, err := os.Create("overview_gas.png")
	if err != nil {
		log.Fatal(err)
	}

	graph.Title = "Gas (gas)"
	graph.Bars = barsG
	if err := graph.Render(chart.PNG, fg); err != nil {
		log.Fatal(err)
	}

	fc, err := os.Create("overview_cost.png")
	if err != nil {
		log.Fatal(err)
	}

	graph.Title = "Cost (gas/us)"
	graph.Bars = barsC
	if err := graph.Render(chart.PNG, fc); err != nil {
		log.Fatal(err)
	}
	os.RemoveAll("mvccdb")
}

func main() {
	getOverview()
	getOpDetail()
}
