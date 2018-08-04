package new_vm

import (
	"testing"

	"context"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

func Init(t *testing.T) (*Monitor, *MockVM, *database.MockIMultiValue, *database.Visitor) {
	mc := NewController(t)
	defer mc.Finish()
	vm := NewMockVM(mc)
	db := database.NewMockIMultiValue(mc)
	vi := database.NewVisitor(100, db)
	pm := NewMonitor()
	pm.vms[""] = vm
	return pm, vm, db, vi
}

func TestMonitor_Call(t *testing.T) {
	monitor, vm, db, vi := Init(t)

	ctx := context.WithValue(context.Background(), "gas_price", uint64(1))

	host := NewHost(ctx, vi)

	vm.EXPECT().LoadAndCall(Any(), Any(), Any(), Any()).DoAndReturn(func(host *Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {

		return []string{"world"}, cost, nil
	})
	db.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (string, error) {
		return "contract", nil
	})

	monitor.Call(host, "contract", "api", "1")
}

func TestMonitor_Context(t *testing.T) {
	monitor, vm, db, vi := Init(t)
	ctx := context.WithValue(context.Background(), "gas_price", uint64(1))

	host := NewHost(ctx, vi)

	outerFlag := false
	innerFlag := false

	vm.EXPECT().LoadAndCall(Any(), Any(), "outer", Any()).DoAndReturn(func(host *Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		outerFlag = true
		monitor.Call(host, "contract", "inner", "hello")

		return []string{"world"}, cost, nil
	})

	vm.EXPECT().LoadAndCall(Any(), Any(), "inner", Any()).DoAndReturn(func(host *Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		innerFlag = true
		return []string{"world"}, cost, nil
	})
	db.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (string, error) {
		return "contract", nil
	})

	monitor.Call(host, "contract", "outer", "1")

	if !outerFlag || !innerFlag {
		t.Fatal(outerFlag, innerFlag)
	}
}

func TestMonitor_HostCall(t *testing.T) {
	monitor, vm, db, vi := Init(t)
	staticMonitor = monitor

	ctx := context.WithValue(context.Background(), "gas_price", uint64(1))
	ctx = context.WithValue(ctx, "stack_height", 1)
	ctx = context.WithValue(ctx, "stack0", "test")

	host := NewHost(ctx, vi)
	outerFlag := false
	innerFlag := false

	vm.EXPECT().LoadAndCall(Any(), Any(), "outer", Any()).DoAndReturn(func(host *Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		outerFlag = true
		host.Call("contract", "inner", "hello")

		return []string{"world"}, cost, nil
	})

	vm.EXPECT().LoadAndCall(Any(), Any(), "inner", Any()).DoAndReturn(func(host *Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		innerFlag = true
		if host.ctx.Value("abi_name") != "inner" {
			t.Fatal(host.ctx)
		}

		return []string{"world"}, cost, nil
	})
	db.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (string, error) {
		return "contract", nil
	})

	monitor.Call(host, "contract", "outer", "1")

	if !outerFlag || !innerFlag {
		t.Fatal(outerFlag, innerFlag)
	}
}
