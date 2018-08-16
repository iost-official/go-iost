package native_vm

import (
	"testing"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
	"os"
	"io/ioutil"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
)

var testDataPath = "./test_data/"

func MyInit(t *testing.T, conName string, optional ...interface{}) (*VM, *host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath(testDataPath + conName + ".json")
	vi := database.NewVisitor(100, db)

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", uint64(1))
	var gasLimit = uint64(10000);
	if len(optional) > 0 {
		gasLimit = optional[0].(uint64)
	}
	ctx.GSet("gas_limit", gasLimit)
	ctx.Set("contract_name", conName)

	pm := new_vm.NewMonitor()
	h := host.NewHost(ctx, vi, pm, nil)

	code := &contract.Contract{
		ID:   conName,
	}

	e := &VM{}
	e.Init()

	return e, h, code
}

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

func TestEngine_SetCode(t *testing.T) {
	e, host, code := MyInit(t, "setcode")
	host.Ctx.Set("tx_hash", []byte("iamhash"))

	rawCode, err := ReadFile(testDataPath + "test.js")
	if err != nil {
		t.Fatalf("read file error: %v\n", err)
	}
	rawAbi, err := ReadFile(testDataPath + "test.js.abi")
	if err != nil {
		t.Fatalf("read file error: %v\n", err)
	}

	compiler := &contract.Compiler{}
	con, err := compiler.Parse("", string(rawCode), string(rawAbi))
	if err != nil {
		t.Fatalf("compiler parse error: %v\n", err)
	}

	rs, _, err := e.LoadAndCall(host, code, "SetCode", con.Encode())

	if err != nil {
		t.Fatalf("LoadAndCall setcode error: %v\n", err)
	}
	if len(rs) != 1 || rs[0].(string) != "0.0000000000800029" {
		t.Errorf("LoadAndCall except 0.0000000000800029, got %s\n", rs[0])
	}
}
