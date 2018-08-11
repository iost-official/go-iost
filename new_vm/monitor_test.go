package new_vm

import (
	"testing"

	"context"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/host"
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

	h := host.NewHost(ctx, vi, monitor)

	flag := false

	vm.EXPECT().LoadAndCall(Any(), Any(), Any(), Any()).DoAndReturn(func(h *host.Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		flag = true
		return []string{"world"}, cost, nil
	})

	c := contract.Contract{
		ID:   "contract",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				&contract.ABI{
					Name:     "abi",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
			},
		},
	}

	db.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	monitor.Call(h, "contract", "abi", "1")

	if !flag {
		t.Fatal(flag)
	}
}

func TestMonitor_Context(t *testing.T) {
	monitor, vm, db, vi := Init(t)
	ctx := context.WithValue(context.Background(), "gas_price", uint64(1))

	h := host.NewHost(ctx, vi, monitor)

	outerFlag := false
	innerFlag := false

	vm.EXPECT().LoadAndCall(Any(), Any(), "outer", Any()).DoAndReturn(func(h *host.Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		outerFlag = true
		monitor.Call(h, "contract", "inner", "hello")

		return []string{"world"}, cost, nil
	})

	vm.EXPECT().LoadAndCall(Any(), Any(), "inner", Any()).DoAndReturn(func(h *host.Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		innerFlag = true
		return []string{"world"}, cost, nil
	})
	c := contract.Contract{
		ID:   "contract",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				&contract.ABI{
					Name:     "outer",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				&contract.ABI{
					Name:     "inner",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
			},
		},
	}

	db.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	monitor.Call(h, "contract", "outer", "1")

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

	h := host.NewHost(ctx, vi, monitor)
	outerFlag := false
	innerFlag := false

	vm.EXPECT().LoadAndCall(Any(), Any(), "outer", Any()).DoAndReturn(func(h *host.Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		outerFlag = true
		h.Call("contract", "inner", "hello")

		return []string{"world"}, cost, nil
	})

	vm.EXPECT().LoadAndCall(Any(), Any(), "inner", Any()).DoAndReturn(func(h *host.Host, c *contract.Contract, api string, args ...string) (rtn []string, cost *contract.Cost, err error) {
		innerFlag = true
		if h.Ctx.Value("abi_name") != "inner" {
			t.Fatal(h.Ctx)
		}

		return []string{"world"}, cost, nil
	})
	c := contract.Contract{
		ID:   "contract",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				&contract.ABI{
					Name:     "outer",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
				&contract.ABI{
					Name:     "inner",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
			},
		},
	}

	db.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	monitor.Call(h, "contract", "outer", "1")

	if !outerFlag || !innerFlag {
		t.Fatal(outerFlag, innerFlag)
	}
}
