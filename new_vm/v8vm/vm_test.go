package v8

import (
	"context"
	"testing"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
)

func Init(t *testing.T) *database.Visitor {
	mc := NewController(t)
	defer mc.Finish()
	db := database.NewMockIMultiValue(mc)
	vi := database.NewVisitor(100, db)
	return vi
}

func TestEngine_LoadAndCall(t *testing.T) {
	vi := Init(t)
	ctx := context.Background()
	ctx = context.WithValue(context.Background(), "gas_price", uint64(1))
	ctx = context.WithValue(ctx, "contract_name", "contractName")
	host := new_vm.NewHost(ctx, vi)

	code := &contract.Contract{
		ContractInfo: contract.ContractInfo{
			Name: "test.js",
		},
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


	rs, err := e.LoadAndCall(host, code, "fibonacci", "12")

	if err != nil {
		t.Fatalf("LoadAndCall run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0] != "144" {
		t.Errorf("LoadAndCall except 144, got %s\n", rs[0])
	}
}
