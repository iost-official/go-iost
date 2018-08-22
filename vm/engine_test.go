package vm

import (
	"testing"

	"time"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	blk "github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

var jsPath = "./v8vm/v8/libjs/"

func engineinit(t *testing.T) (*blk.BlockHead, *database.MockIMultiValue, *MockVM) {
	ctl := gomock.NewController(t)
	db := database.NewMockIMultiValue(ctl)
	bh := &blk.BlockHead{
		ParentHash: []byte("abc"),
		Number:     10,
		Witness:    "witness",
		Time:       123456,
	}
	vm := NewMockVM(ctl)
	pm := NewMonitor()
	pm.vms["mock"] = vm

	//nvm := native.VM{}
	//
	//pm.vms["native"] = &nvm
	staticMonitor = pm
	return bh, db, vm
}

func TestNewEngine(t *testing.T) { // test of normal engine work
	bh, db, vm := engineinit(t)
	e := NewEngine(bh, db)
	e.SetUp("js_path", jsPath)

	bi, _ := e.(*engineImpl).ho.BlockInfo()

	blkInfo := string(bi)
	if blkInfo != `{"number":10,"parent_hash":"ZiCa","time":123456,"witness":"witness"}` {
		t.Fatal(blkInfo)
	}

	//ac, err := account.NewAccount(nil)
	//if err != nil {
	//	t.Fatal(err)
	//}

	mtx := tx.Tx{
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	act := tx.Action{
		Contract:   "contract",
		ActionName: "abi",
		Data:       `["datas"]`,
	}

	mtx.Actions = append(mtx.Actions, act)

	c := contract.Contract{
		ID:   "contract",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "mock",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "abi",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(10),
					Limit:    contract.NewCost(100, 100, 100),
				},
			},
		},
	}

	db.EXPECT().Get("state", "c-contract").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000000)), nil
	})

	vm.EXPECT().LoadAndCall(gomock.Any(), gomock.Any(), "abi", `datas`).DoAndReturn(
		func(host *host.Host, c *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			return nil, contract.Cost0(), nil
		},
	)

	committed := false

	db.EXPECT().Commit().Do(func() {
		committed = true
	})

	txr, err := e.Exec(&mtx)
	if err != nil {
		t.Fatal(err)
	}
	if txr.Status.Code != tx.Success {
		t.Fatal(txr.Status)
	}

	if !committed {
		t.Fatal(committed)
	}
}

func TestLogger(t *testing.T) { // test of normal engine work
	bh, db, vm := engineinit(t)
	e := NewEngine(bh, db)
	e.SetUp("js_path", jsPath)
	e.SetUp("log_level", "debug")
	e.SetUp("log_enable", "")

	mtx := tx.Tx{
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	act := tx.Action{
		Contract:   "contract",
		ActionName: "abi",
		Data:       `["datas"]`,
	}

	mtx.Actions = append(mtx.Actions, act)

	c := contract.Contract{
		ID:   "contract",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "mock",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "abi",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(10),
					Limit:    contract.NewCost(100, 100, 100),
				},
			},
		},
	}

	db.EXPECT().Get("state", "c-contract").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000000)), nil
	})

	vm.EXPECT().LoadAndCall(gomock.Any(), gomock.Any(), "abi", `datas`).DoAndReturn(
		func(host *host.Host, c *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			l := host.Logger()
			l.Error("test of error")
			l.Debug("test of debug")
			l.Info("test of info")
			return nil, contract.Cost0(), nil
		},
	)

	committed := false

	db.EXPECT().Commit().Do(func() {
		committed = true
	})

	txr, err := e.Exec(&mtx)
	if err != nil {
		t.Fatal(err)
	}
	if txr.Status.Code != tx.Success {
		t.Fatal(txr.Status)
	}

	if !committed {
		t.Fatal(committed)
	}
}

func TestCost(t *testing.T) { // tests of context transport
	bh, db, vm := engineinit(t)
	e := NewEngine(bh, db)
	e.SetUp("js_path", jsPath)

	mtx := tx.Tx{
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "contract",
		ActionName: "abi",
		Data:       `["datas"]`,
	}

	mtx.Actions = append(mtx.Actions, ac)

	ac2 := tx.Action{
		Contract:   "contract",
		ActionName: "abi",
		Data:       `["data2"]`,
	}
	mtx.Actions = append(mtx.Actions, ac2)

	c := contract.Contract{
		ID:   "contract",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "mock",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "abi",
					Args:     []string{"string"},
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
				},
			},
		},
	}

	db.EXPECT().Get("state", "c-contract").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000000)), nil
	})

	db.EXPECT().Get("state", "i-CGcontract").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-witness").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value) != int64(999900) {
			t.Fatal(database.MustUnmarshal(value))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-witness", gomock.Any()).Times(2).DoAndReturn(func(table string, key string, value string) error {

		//fmt.Println("witness received money", database.MustUnmarshal(value))
		if database.MustUnmarshal(value) != int64(1100) {
			t.Fatal(database.MustUnmarshal(value))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-CGcontract", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value) != int64(900) {
			t.Fatal(database.MustUnmarshal(value))
		}
		return nil
	})

	vm.EXPECT().LoadAndCall(gomock.Any(), gomock.Any(), "abi", "datas").DoAndReturn(
		func(host *host.Host, c *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.NewCost(0, 0, 100)
			//fmt.Println("call datas")
			return nil, cost, nil
		},
	)

	vm.EXPECT().LoadAndCall(gomock.Any(), gomock.Any(), "abi", "data2").DoAndReturn(
		func(host *host.Host, c *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			host.ABIConfig("payment", "contract_pay")
			cost = contract.NewCost(0, 0, 100)
			//fmt.Println("call data2")
			return nil, cost, nil
		},
	)

	committed := false

	db.EXPECT().Commit().Do(func() {
		committed = true
	})

	txr, err := e.Exec(&mtx)
	if err != nil {
		t.Fatal(err)
	}
	if txr.Status.Code != tx.Success {
		t.Fatal(txr.Status)
	}

	if !committed {
		t.Fatal(committed)
	}
}

func TestNative_Transfer(t *testing.T) { // tests of native vm works
	bh, db, _ := engineinit(t)
	e := NewEngine(bh, db)
	e.SetUp("js_path", jsPath)
	mtx := tx.Tx{
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
		Signers:    make([][]byte, 0),
	}
	mtx.Signers = append(mtx.Signers, []byte("a"))

	ac := tx.Action{
		Contract:   "iost.system",
		ActionName: "Transfer",
		Data:       `["a","b", 100]`,
	}

	mtx.Actions = append(mtx.Actions, ac)

	c := contract.Contract{
		ID:   "iost.system",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "native",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "Transfer",
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
					Args:     []string{"string", "string", "number"},
				},
			},
		},
	}

	db.EXPECT().Get("state", "c-iost.system").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-a").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-witness").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(10000000)), nil
	})

	db.EXPECT().Put("state", "i-a", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(900) {
			t.Fatal("a", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1100) {
			t.Fatal("b", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-witness", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1003) {
			t.Fatal("witness", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(9999997) {
			t.Fatal("publisher", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	committed := false

	db.EXPECT().Commit().Do(func() {
		committed = true
	})

	txr, err := e.Exec(&mtx)
	if err != nil {
		t.Fatal(err)
	}
	if txr.Status.Code != tx.Success {
		t.Fatal(txr.Status)
	}

	if !committed {
		t.Fatal(committed)
	}

}

func TestNative_TopUp(t *testing.T) { // tests of native vm works
	bh, db, _ := engineinit(t)
	e := NewEngine(bh, db)
	e.SetUp("js_path", jsPath)

	mtx := tx.Tx{
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "iost.system",
		ActionName: "TopUp",
		Data:       `["a","b",100]`,
	}

	mtx.Actions = append(mtx.Actions, ac)

	c := contract.Contract{
		ID:   "iost.system",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "native",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "TopUp",
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
					Args:     []string{"string", "string", "number"},
				},
			},
		},
	}

	db.EXPECT().Get("state", "c-iost.system").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-CGa").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Put("state", "i-CGa", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1100) {
			t.Fatal("CGa", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(900) {
			t.Fatal("a", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Get("state", "i-witness").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000000)), nil
	})

	db.EXPECT().Put("state", "i-witness", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1003) {
			t.Fatal("witness", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(999997) {
			t.Fatal("publisher", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	committed := false

	db.EXPECT().Commit().Do(func() {
		committed = true
	})

	txr, err := e.Exec(&mtx)
	if err != nil {
		t.Fatal(err)
	}
	if txr.Status.Code != tx.Success {
		t.Fatal(txr.Status)
	}

	if !committed {
		t.Fatal(committed)
	}

}

func TestNative_Receipt(t *testing.T) { // tests of native vm works
	bh, db, _ := engineinit(t)
	e := NewEngine(bh, db)
	e.SetUp("js_path", jsPath)
	mtx := tx.Tx{
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "iost.system",
		ActionName: "CallWithReceipt",
		Data:       `["iost.system", "Receipt", ["iamreceipt"]]`,
	}

	mtx.Actions = append(mtx.Actions, ac)

	c := contract.Contract{
		ID:   "iost.system",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "native",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "Receipt",
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
					Args:     []string{"string"},
				},
				{
					Name:     "CallWithReceipt",
					Payment:  0,
					GasPrice: int64(1000),
					Limit:    contract.NewCost(100, 100, 100),
					Args:     []string{"string", "string", "json"},
				},
			},
		},
	}

	db.EXPECT().Get("state", "c-iost.system").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-witness").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(10000000)), nil
	})

	db.EXPECT().Put("state", "i-witness", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1103) {
			t.Fatal("witness", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(9999897) {
			t.Fatal("publisher", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	committed := false

	db.EXPECT().Commit().Do(func() {
		committed = true
	})

	txr, err := e.Exec(&mtx)
	if err != nil {
		t.Fatal(err)
	}
	if txr.Status.Code != tx.Success {
		t.Fatal(txr.Status)
	}
	if len(txr.Receipts) != 2 || txr.Receipts[0].Type != tx.UserDefined || txr.Receipts[0].Content != "iamreceipt" ||
		txr.Receipts[1].Type != tx.SystemDefined || txr.Receipts[1].Content != `["Receipt",["iamreceipt"],"success"]` {
		t.Fatal(txr.Receipts)
	}

	if !committed {
		t.Fatal(committed)
	}
}

func TestJS(t *testing.T) {
	bh, db, _ := engineinit(t)
	e := NewEngine(bh, db)
	e.SetUp("js_path", jsPath)

	mtx := tx.Tx{
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "testjs",
		ActionName: "hello",
		Data:       `[]`,
	}

	mtx.Actions = append(mtx.Actions, ac)

	c := contract.Contract{
		ID: "testjs",
		Code: `
class Contract {
 constructor() {
  
 }
 hello() {
  return "show";
 }
}

module.exports = Contract;
`,
		Info: &contract.Info{
			Lang:        "javascript",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "hello",
					Payment:  0,
					GasPrice: int64(1),
					Limit:    contract.NewCost(100, 100, 100),
					Args:     []string{},
				},
			},
		},
	}

	db.EXPECT().Get("state", "c-testjs").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000000)), nil
	})

	db.EXPECT().Get("state", "i-witness").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {

		if database.MustUnmarshal(value).(int64) != 999993 {
			t.Fatal(database.MustUnmarshal(value).(int64))
		}

		return nil
	})

	db.EXPECT().Put("state", "i-witness", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1007) {
			t.Fatal("witness", database.MustUnmarshal(value).(int64))
		}
		return nil
	})
	db.EXPECT().Rollback().Do(func() {
		t.Log("exec tx failed, and success rollback")
	})

	db.EXPECT().Commit().Do(func() {
		t.Log("committed")
	})

	txr, err := e.Exec(&mtx)
	if err != nil {
		t.Fatal(err)
	}
	if txr.Status.Code != tx.Success {
		t.Fatal(txr.Status)
	}
}
