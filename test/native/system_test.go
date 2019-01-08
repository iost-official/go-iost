package native

import (
	"io/ioutil"
	"testing"

	"time"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/vm/native"
)

var testDataPath = "./test_data/"

func InitVMWithMonitor(t *testing.T, conName string, optional ...interface{}) (*native.Impl, *host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath(testDataPath + conName + ".json")
	vi := database.NewVisitor(100, db)

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", int64(1))
	var gasLimit = int64(100000)
	if len(optional) > 0 {
		gasLimit = optional[0].(int64)
	}
	ctx.GSet("gas_limit", gasLimit)
	ctx.Set("contract_name", conName)
	ctx.Set("tx_hash", []byte("iamhash"))
	ctx.Set("auth_list", make(map[string]int))
	ctx.Set("publisher", "pub")

	pm := vm.NewMonitor()
	h := host.NewHost(ctx, vi, pm, nil)
	h.Context().Set("stack_height", 0)

	code := &contract.Contract{
		ID: "system.iost",
	}

	e := &native.Impl{}
	e.Init()

	return e, h, code
}

// nolint
func TestEngine_SetCode(t *testing.T) {

	e, host, code := InitVMWithMonitor(t, "setcode")
	host.Context().Set("tx_hash", "iamhash")
	host.Context().Set("contract_name", "system.iost")
	host.Context().Set("auth_contract_list", make(map[string]int))
	host.SetDeadline(time.Now().Add(10 * time.Second))
	hash := "Contractiamhash"

	rawCode, err := ioutil.ReadFile(testDataPath + "test.js")
	if err != nil {
		t.Fatalf("read file error: %v\n", err)
	}
	rawAbi, err := ioutil.ReadFile(testDataPath + "test.js.abi")
	if err != nil {
		t.Fatalf("read file error: %v\n", err)
	}

	compiler := &contract.Compiler{}
	con, err := compiler.Parse("", string(rawCode), string(rawAbi))
	if err != nil {
		t.Fatalf("compiler parse error: %v\n", err)
	}

	rs, _, err := e.LoadAndCall(host, code, "SetCode", con.B64Encode())

	if err != nil {
		t.Fatalf("LoadAndCall setcode error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != hash {
		t.Fatalf("LoadAndCall except Contract"+"iamhash"+", got %s\n", rs[0])
	}

	con.ID = "Contractiamhash"

	rawCode, err = ioutil.ReadFile(testDataPath + "test_new.js")
	if err != nil {
		t.Fatalf("read file error: %v\n", err)
	}
	rawAbi, err = ioutil.ReadFile(testDataPath + "test_new.js.abi")
	if err != nil {
		t.Fatalf("read file error: %v\n", err)
	}
	con, err = compiler.Parse(con.ID, string(rawCode), string(rawAbi))
	if err != nil {
		t.Fatalf("compiler parse error: %v\n", err)
	}

	rs, _, err = e.LoadAndCall(host, code, "UpdateCode", con.B64Encode(), "")
	if err != nil {
		t.Fatalf("LoadAndCall update error: %v\n", err)
	}
	if len(rs) != 0 {
		t.Fatalf("LoadAndCall except 0 rtn"+", got %d\n", len(rs))
	}
}
