package v8

import (
	"context"
	"testing"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
)

func TestEngine_LoadAndCall(t *testing.T) {
	contract := &new_vm.Contract{
		ContractInfo: new_vm.ContractInfo{
			Name: "test.js",
		},
		Code: `var Contract = function() {
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

	rs, err := e.LoadAndCall(context.Background(), contract, "fibonacci", "12")

	if err != nil {
		t.Fatalf("LoadAndCall run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0] != "144" {
		t.Errorf("LoadAndCall except 144, got %s\n", rs[0])
	}
}
