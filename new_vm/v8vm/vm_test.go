package v8

import (
	"context"
	"testing"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
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

	rs, _, err := e.LoadAndCall(host, code, "fibonacci", "12")

	if err != nil {
		t.Fatalf("LoadAndCall run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0] != "144" {
		t.Errorf("LoadAndCall except 144, got %s\n", rs[0])
	}
}

//func TestEngine_Storage(t *testing.T) {
//	vi := Init(t)
//	ctx := context.Background()
//	ctx = context.WithValue(context.Background(), "gas_price", uint64(1))
//	ctx = context.WithValue(ctx, "contract_name", "contractName")
//	host := new_vm.NewHost(ctx, vi)
//
//	code := &contract.Contract{
//		ID: "test.js",
//		Code: `
//var Contract = function() {
//};
//
//	Contract.prototype = {
//	mySet: function(k, v) {
//			return IOSTContractStorage.put(k, v);
//		},
//	myGet: function(k) {
//			return IOSTContractStorage.get(k)
//		}
//	};
//
//	module.exports = Contract;
//`,
//	}
//
//	e := NewVM()
//	defer e.Release()
//	e.Init()
//
//
//	e.LoadAndCall(host, code, "mySet", "mySetKey", "mySetVal")
//	rs, _,err := e.LoadAndCall(host, code, "myGet", "mySetKey")
//
//	if err != nil {
//		t.Fatalf("LoadAndCall run error: %v\n", err)
//	}
//	if len(rs) != 1 || rs[0] != "mySetVal" {
//		t.Errorf("LoadAndCall except mySetVal, got %s\n", rs[0])
//	}
//}

func TestEngine_bigNumber(t *testing.T) {
	vi := Init(t)
	ctx := context.Background()
	ctx = context.WithValue(context.Background(), "gas_price", uint64(1))
	ctx = context.WithValue(ctx, "contract_name", "contractName")
	host := new_vm.NewHost(ctx, vi)

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

	//e.LoadAndCall(host, code, "mySet", "mySetKey", "mySetVal")
	rs, _, err := e.LoadAndCall(host, code, "getVal")

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
	ctx = context.WithValue(ctx, "contract_name", "contractName")
	host := new_vm.NewHost(ctx, vi)

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

	//e.LoadAndCall(host, code, "mySet", "mySetKey", "mySetVal")
	_, _, err := e.LoadAndCall(host, code, "loop")

	if err != nil && err.Error() != "execution killed" {
		t.Fatalf("infiniteLoop run error: %v\n", err)
	}
}

func TestEngine_injectCode(t *testing.T) {
	vi := Init(t)
	ctx := context.Background()
	ctx = context.WithValue(context.Background(), "gas_price", uint64(1))
	ctx = context.WithValue(ctx, "contract_name", "contractName")
	host := new_vm.NewHost(ctx, vi)

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
		},
		callFib: function(cycles) {
			var result = this.fibonacci(cycles);
			var gasCount = _IOSTInstruction_counter.count();
			return gasCount;
		}
	}

	module.exports = Contract
`,
	}

	e := NewVM()
	defer e.Release()
	e.Init()

	rs, _, err := e.LoadAndCall(host, code, "callFib", "12")

	if err != nil {
		t.Fatalf("LoadAndCall run error: %v\n", err)
	}
	if len(rs) != 1 || rs[0] != "10217" {
		t.Errorf("LoadAndCall except 144, got %s\n", rs[0])
	}
}
