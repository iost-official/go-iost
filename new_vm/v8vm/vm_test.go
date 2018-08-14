package v8

import (
	"testing"

	"io/ioutil"
	"os"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
	"strings"
)

var testDataPath = "./test_data/"

func ReadFile(src string) ([]byte, error) {
	fi, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func Init(t *testing.T) *database.Visitor {
	mc := NewController(t)
	defer mc.Finish()
	db := database.NewMockIMultiValue(mc)
	vi := database.NewVisitor(100, db)
	return vi
}

func MyInit(t *testing.T, conName string, optional ...interface{}) (*VM, *host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath(testDataPath + conName + ".json")
	vi := database.NewVisitor(100, db)

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", uint64(1))
	var gasLimit = uint64(10000);
	if len(optional) > 0 {
		gasLimit = optional[0].(uint64)
	}
	ctx.Set("gas_limit", gasLimit)
	ctx.Set("contract_name", conName)
	h := &host.Host{Ctx: ctx, DB: vi}

	fd, err := ReadFile(testDataPath + conName + ".js")
	if err != nil {
		t.Fatal("Read file failed: ", err.Error())
		return nil, nil, nil
	}
	rawCode := string(fd)

	code := &contract.Contract{
		ID:   conName,
		Code: rawCode,
	}

	e := NewVM()
	e.Init()
	e.SetJSPath("./v8/libjs/")

	return e, h, code
}

func TestEngine_LoadAndCall(t *testing.T) {
	vi := Init(t)
	ctx := host.NewContext(nil)
	ctx.Set("gas_price", uint64(1))
	ctx.Set("gas_limit", uint64(1000000000))
	ctx.Set("contract_name", "contractName")
	tHost := &host.Host{Ctx: ctx, DB: vi}

	code := &contract.Contract{
		ID: "test.js",
		Code: `
var Contract = function() {
}

	Contract.prototype = {
	fibonacci: function(cycles) {
			if (cycles == 0) return 0;
			if (cycles == 1) return 1;
			return this.fibonacci(cycles - 1) + this.fibonacci(cycles - 2);
		}
	}

	module.exports = Contract
`,
	}

	e := NewVM()
	defer e.Release()
	e.Init()
	e.SetJSPath("./v8/libjs/")

	rs, _, err := e.LoadAndCall(tHost, code, "fibonacci", "12")

	if err != nil {
		t.Fatalf("LoadAndCall run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0] != "144" {
		t.Errorf("LoadAndCall except 144, got %s\n", rs[0])
	}
}

func TestEngine_bigNumber(t *testing.T) {
	vi := Init(t)
	ctx := host.NewContext(nil)
	ctx.Set("gas_price", uint64(1))
	ctx.Set("gas_limit", uint64(1000000000))
	ctx.Set("contract_name", "contractName")
	tHost := &host.Host{Ctx: ctx, DB: vi}

	code := &contract.Contract{
		ID: "test.js",
		Code: `
var Contract = function() {
	this.val = new BigNumber(0.00000000008);
	this.val = this.val.plus(0.0000000000000029);
};

	Contract.prototype = {
	getVal: function() {
		return this.val.toString(10);
	}
	};

	module.exports = Contract;
`,
	}

	e := NewVM()
	defer e.Release()
	e.Init()
	e.SetJSPath("./v8/libjs/")

	//e.LoadAndCall(host, code, "mySet", "mySetKey", "mySetVal")
	e.LoadAndCall(tHost, code, "constructor")
	rs, _, err := e.LoadAndCall(tHost, code, "getVal")

	if err != nil {
		t.Fatalf("LoadAndCall run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0] != "0.0000000000800029" {
		t.Errorf("LoadAndCall except mySetVal, got %s\n", rs[0])
	}
}

func TestEngine_infiniteLoop(t *testing.T) {
	vi := Init(t)

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", uint64(1))
	ctx.Set("gas_limit", uint64(1000))
	ctx.Set("contract_name", "contractName")
	tHost := &host.Host{Ctx: ctx, DB: vi}

	code := &contract.Contract{
		ID: "test.js",
		Code: `
var Contract = function() {
};

	Contract.prototype = {
	loop: function() {
		while (true) {}
	}
	};

	module.exports = Contract;
`,
	}

	e := NewVM()
	defer e.Release()
	e.Init()
	e.SetJSPath("./v8/libjs/")

	//e.LoadAndCall(host, code, "mySet", "mySetKey", "mySetVal")
	_, _, err := e.LoadAndCall(tHost, code, "loop")

	if err != nil && err.Error() != "out of gas" {
		t.Fatalf("infiniteLoop run error: %v\n", err)
	}
}

func TestEngine_Storage(t *testing.T) {
	e, host, code := MyInit(t, "storage1")

	rs, _, err := e.LoadAndCall(host, code, "get", "a")
	if err != nil {
		t.Fatalf("LoadAndCall get run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "1000" {
		t.Errorf("LoadAndCall except mySetVal, got %s\n", rs[0])
	}

	rtn, cost, err := e.LoadAndCall(host, code, "put", "mySetKey", "mySetVal")
	if err != nil {
		t.Fatalf("LoadAndCall put run error: %v\n", err)
	}
	if len(rtn) != 1 || rtn[0].(string) != "0" {
		t.Fatalf("return of put should be ['0']")
	}
	t.Log(cost)

	rs, _, err = e.LoadAndCall(host, code, "get", "mySetKey")
	if err != nil {
		t.Fatalf("LoadAndCall get run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "mySetVal" {
		t.Errorf("LoadAndCall except mySetVal, got %s\n", rs[0])
	}

	rtn, cost, err = e.LoadAndCall(host, code, "delete", "mySetKey")
	if err != nil {
		t.Fatalf("LoadAndCall delete run error: %v\n", err)
	}
	if len(rtn) != 1 || rtn[0].(string) != "0" {
		t.Fatalf("return of put should be ['0']")
	}
	t.Log(cost)

	rs, _, err = e.LoadAndCall(host, code, "get", "mySetKey")
	if err != nil {
		t.Fatalf("LoadAndCall get run error: %v\n", err)
	}
	// todo get return nil
	if len(rs) != 1 || rs[0].(string) != "nil" {
		t.Errorf("LoadAndCall except mySetVal, got %s\n", rs[0])
	}
}

func TestEngine_DataType(t *testing.T) {
	e, host, code := MyInit(t, "datatype")

	rs, _,err := e.LoadAndCall(host, code, "number", 1)
	if err != nil {
		t.Fatalf("LoadAndCall number run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "0.5555555555" {
		t.Errorf("LoadAndCall except 0.5555555555, got %s\n", rs[0])
	}

	rs, _,err = e.LoadAndCall(host, code, "number_big", 1)
	if err != nil {
		t.Fatalf("LoadAndCall number_big run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "0.5555555555" {
		t.Errorf("LoadAndCall except 0.5555555555, got %s\n", rs[0])
	}

	rs, _,err = e.LoadAndCall(host, code, "number_op")
	if err != nil {
		t.Fatalf("LoadAndCall number_op run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "3" {
		t.Errorf("LoadAndCall except 3, got %s\n", rs[0])
	}

	rs, _,err = e.LoadAndCall(host, code, "number_op2")
	if err != nil {
		t.Fatalf("LoadAndCall number_op2 run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "2" {
		t.Errorf("LoadAndCall except 3, got %s\n", rs[0])
	}

	rs, _,err = e.LoadAndCall(host, code, "number_strange")
	if err != nil {
		t.Fatalf("LoadAndCall number_strange run error: %v\n", err)
	}
	// todo get return string -infinity
	if len(rs) != 1 || rs[0].(string) != "-Infinity" {
		t.Errorf("LoadAndCall except Infinity, got %s\n", rs[0])
	}

	rs, _,err = e.LoadAndCall(host, code, "param", 2)
	if err != nil {
		t.Fatalf("LoadAndCall param run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "4" {
		t.Errorf("LoadAndCall except 4, got %s %s\n", rs[0])
	}

	rs, _,err = e.LoadAndCall(host, code, "param2")
	if err != nil {
		t.Fatalf("LoadAndCall param run error: %v\n", err)
	}
	// todo get return string undefined
	if len(rs) != 1 || rs[0].(string) != "undefined" {
		t.Errorf("LoadAndCall except undefined, got %s\n", rs[0])
	}

	rs, _,err = e.LoadAndCall(host, code, "bool")
	if err != nil {
		t.Fatalf("LoadAndCall bool run error: %v\n", err)
	}
	// todo get return string false
	if len(rs) != 1 || rs[0].(string) != "false" {
		t.Errorf("LoadAndCall except undefined, got %s\n", rs[0])
	}

	rs, _,err = e.LoadAndCall(host, code, "string")
	if err != nil {
		t.Fatalf("LoadAndCall string run error: %v\n", err)
	}
	if len(rs) != 1 || len(rs[0].(string)) != 4096 {
		t.Errorf("LoadAndCall except len 4096, got %d\n", len(rs[0].(string)))
	}

	rs, _,err = e.LoadAndCall(host, code, "array")
	if err != nil {
		t.Fatalf("LoadAndCall array run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "0,1,2,3"{
		t.Errorf("LoadAndCall except 0,1,2,3, got %s\n", rs[0].(string))
	}

	// todo get return string [object Object]
	/*
	rs, _,err = e.LoadAndCall(host, code, "object")
	if err != nil {
		t.Fatalf("LoadAndCall array run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "0,1,2,3"{
		t.Errorf("LoadAndCall except 0,1,2,3, got %s\n", rs[0].(string))
	}
	*/
}

func TestEngine_Loop(t *testing.T) {
	e, host, code := MyInit(t, "loop")

	_, _,err := e.LoadAndCall(host, code, "for")
	if err == nil || err.Error() != "out of gas"{
		t.Fatalf("LoadAndCall for should return error: out of gas, but got %v\n", err)
	}

	e, host, code = MyInit(t, "loop")
	_, _,err = e.LoadAndCall(host, code, "for2")
	if err == nil || err.Error() != "out of gas"{
		t.Fatalf("LoadAndCall for should return error: out of gas, but got %v\n", err)
	}

	e, host, code = MyInit(t, "loop")
	rs, _,err := e.LoadAndCall(host, code, "for3")
	if err != nil {
		t.Fatalf("LoadAndCall for3 run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "10"{
		t.Errorf("LoadAndCall except 10, got %s\n", rs[0].(string))
	}

	e, host, code = MyInit(t, "loop")
	rs, _,err = e.LoadAndCall(host, code, "forin")
	if err != nil {
		t.Fatalf("LoadAndCall forin run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "12"{
		t.Errorf("LoadAndCall except 12, got %s\n", rs[0].(string))
	}

	e, host, code = MyInit(t, "loop")
	rs, _,err = e.LoadAndCall(host, code, "forof")
	if err != nil {
		t.Fatalf("LoadAndCall forof run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "6"{
		t.Errorf("LoadAndCall except 6, got %s\n", rs[0].(string))
	}

	e, host, code = MyInit(t, "loop")
	_, _,err = e.LoadAndCall(host, code, "while")
	if err == nil || err.Error() != "out of gas"{
		t.Fatalf("LoadAndCall for should return error: out of gas, but got %v\n", err)
	}

	e, host, code = MyInit(t, "loop")
	_, _,err = e.LoadAndCall(host, code, "dowhile")
	if err == nil || err.Error() != "out of gas"{
		t.Fatalf("LoadAndCall for should return error: out of gas, but got %v\n", err)
	}
}

func TestEngine_Func(t *testing.T) {
	e, host, code := MyInit(t, "func")
	_, _,err := e.LoadAndCall(host, code, "func1")
	if err == nil || err.Error() != "out of gas"{
		t.Fatalf("LoadAndCall for should return error: out of gas, but got %v\n", err)
	}

	e, host, code = MyInit(t, "func", uint64(100000000000))
	_, _,err = e.LoadAndCall(host, code, "func1")
	if err == nil || !strings.Contains(err.Error(), "Uncaught exception: RangeError: Maximum call stack size exceeded") {
		t.Fatalf("LoadAndCall for should return error: Uncaught exception: RangeError: Maximum call stack size exceeded, but got %v\n", err)
	}
}
