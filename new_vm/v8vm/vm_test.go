package v8

import (
	"context"
	"testing"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
	"os"
	"io/ioutil"
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

func MyInit(t *testing.T, conName string) (*VM, *host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath(testDataPath + conName + ".json")
	vi := database.NewVisitor(100, db)

	ctx := context.Background()
	ctx = context.WithValue(context.Background(), "gas_price", uint64(1))
	ctx = context.WithValue(ctx, "gas_limit", uint64(1000))
	ctx = context.WithValue(ctx, "contract_name", conName)
	host := &host.Host{Ctx: ctx, DB: vi}

	fd, err := ReadFile(testDataPath + conName + ".js")
	if err != nil {
		t.Fatal("Read file failed: ", err.Error())
		return nil, nil, nil
	}
	rawCode := string(fd)

	code := &contract.Contract{
		ID: conName,
		Code: rawCode,
	}

	e := NewVM()
	e.Init()
	e.SetJSPath("./v8/libjs/")

	return e, host, code
}

func TestEngine_LoadAndCall(t *testing.T) {
	vi := Init(t)
	ctx := context.Background()
	ctx = context.WithValue(context.Background(), "gas_price", uint64(1))
	ctx = context.WithValue(ctx, "gas_limit", uint64(1000000000))
	ctx = context.WithValue(ctx, "contract_name", "contractName")
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
	ctx := context.Background()
	ctx = context.WithValue(context.Background(), "gas_price", uint64(1))
	ctx = context.WithValue(ctx, "gas_limit", uint64(1000000000))
	ctx = context.WithValue(ctx, "contract_name", "contractName")
	tHost := &host.Host{Ctx: ctx, DB: vi}

	code := &contract.Contract{
		ID: "test.js",
		Code: `
var Contract = function() {
	this.val = new bigNumber(0.00000000008);
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
	ctx := context.Background()
	ctx = context.WithValue(context.Background(), "gas_price", uint64(1))
	ctx = context.WithValue(ctx, "gas_limit", uint64(1000))
	ctx = context.WithValue(ctx, "contract_name", "contractName")
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

	rs, _,err := e.LoadAndCall(host, code, "get", "a")
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

	rs, _,err = e.LoadAndCall(host, code, "get", "mySetKey")
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

	rs, _,err = e.LoadAndCall(host, code, "get", "mySetKey")
	if err != nil {
		t.Fatalf("LoadAndCall get run error: %v\n", err)
	}
	// todo get return nil
	if len(rs) != 1 || rs[0].(string) != "nil" {
		t.Errorf("LoadAndCall except mySetVal, got %s\n", rs[0])
	}
}

func TestEngine_DataType(t *testing.T) {
}
