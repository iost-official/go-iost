package vm

import (
	"testing"

	"time"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/account"
	blk "github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/event"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

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
	db.EXPECT().Get("state", "c-iost.system").DoAndReturn(func(table string, key string) (string, error) {
		return "-", nil
	})
	db.EXPECT().Get("state", "i-CAiost.bonus-b").DoAndReturn(func(table string, key string) (string, error) {
		return "-", nil
	})
	db.EXPECT().Put("state", "c-iost.system", gomock.Any()).DoAndReturn(func(table string, key string, content string) error {
		return nil
	})
	db.EXPECT().Put("state", "i-CAiost.bonus-b", gomock.Any()).DoAndReturn(func(table string, key string, content string) error {
		return nil
	})
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
		Publisher:  &crypto.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	act := tx.Action{
		Contract:   "Contract0",
		ActionName: "abi",
		Data:       `["datas"]`,
	}

	mtx.Actions = append(mtx.Actions, &act)

	c := contract.Contract{
		ID:   "Contract0",
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

	db.EXPECT().Get("state", "c-Contract0").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b").DoAndReturn(func(table string, key string) (string, error) {
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
		Publisher:  &crypto.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	act := tx.Action{
		Contract:   "Contract0",
		ActionName: "abi",
		Data:       `["datas"]`,
	}

	mtx.Actions = append(mtx.Actions, &act)

	c := contract.Contract{
		ID:   "Contract0",
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

	db.EXPECT().Get("state", "c-Contract0").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b").DoAndReturn(func(table string, key string) (string, error) {
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
		Publisher:  &crypto.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "Contract0",
		ActionName: "abi",
		Data:       `["datas"]`,
	}

	mtx.Actions = append(mtx.Actions, &ac)

	ac2 := tx.Action{
		Contract:   "Contract0",
		ActionName: "abi",
		Data:       `["data2"]`,
	}
	mtx.Actions = append(mtx.Actions, &ac2)

	c := contract.Contract{
		ID:   "Contract0",
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

	db.EXPECT().Get("state", "c-Contract0").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000000)), nil
	})

	db.EXPECT().Get("state", "i-CGContract0-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-witness-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value) != int64(999900) {
			t.Fatal(database.MustUnmarshal(value))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-witness-b", gomock.Any()).Times(2).DoAndReturn(func(table string, key string, value string) error {

		//fmt.Println("witness received money", database.MustUnmarshal(value))
		if database.MustUnmarshal(value) != int64(1090) {
			t.Fatal(database.MustUnmarshal(value))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-CGContract0-b", gomock.Any()).Times(2).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value) != int64(910) && database.MustUnmarshal(value) != int64(900) {
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
		Publisher:  &crypto.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
		Signers:    make([][]byte, 0),
	}
	mtx.Signers = append(mtx.Signers, []byte("a"))

	ac := tx.Action{
		Contract:   "iost.system",
		ActionName: "Transfer",
		Data:       `["a","b", 100]`,
	}

	mtx.Actions = append(mtx.Actions, &ac)

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

	db.EXPECT().Get("state", "i-a-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-b-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-witness-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(10000000)), nil
	})

	db.EXPECT().Put("state", "i-a-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(900) {
			t.Fatal("a", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-b-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1100) {
			t.Fatal("b", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-witness-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1273) {
			t.Fatal("witness", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b", gomock.Any()).Times(2).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(9999727) && database.MustUnmarshal(value).(int64) != int64(9999697) {
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
		Publisher:  &crypto.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "iost.system",
		ActionName: "TopUp",
		Data:       `["a","b",100]`,
	}

	mtx.Actions = append(mtx.Actions, &ac)

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

	db.EXPECT().Get("state", "i-CGa-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-b-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Put("state", "i-CGa-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1100) {
			t.Fatal("CGa", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-b-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(900) {
			t.Fatal("a", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Get("state", "i-witness-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000000)), nil
	})

	db.EXPECT().Put("state", "i-witness-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1273) {
			t.Fatal("witness", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b", gomock.Any()).Times(2).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(999727) && database.MustUnmarshal(value).(int64) != int64(999697) {
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

// nolint
func TestNative_Receipt(t *testing.T) { // tests of native vm works
	bh, db, _ := engineinit(t)
	e := NewEngine(bh, db)
	e.SetUp("js_path", jsPath)
	mtx := tx.Tx{
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  &crypto.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "iost.system",
		ActionName: "CallWithReceipt",
		Data:       `["iost.system", "Receipt", ["iamreceipt"]]`,
	}

	mtx.Actions = append(mtx.Actions, &ac)

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
					GasPrice: int64(100),
					Limit:    contract.NewCost(1000, 1000, 1000),
					Args:     []string{"string"},
				},
				{
					Name:     "CallWithReceipt",
					Payment:  0,
					GasPrice: int64(100),
					Limit:    contract.NewCost(1000, 1000, 1000),
					Args:     []string{"string", "string", "json"},
				},
			},
		},
	}
	db.EXPECT().Get("state", "c-iost.system").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-witness-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(10000000)), nil
	})

	db.EXPECT().Put("state", "i-witness-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1004) {
			t.Fatal("witness", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b", gomock.Any()).Times(2).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(9999996) {
			t.Fatal("publisher", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	committed := false

	db.EXPECT().Commit().Do(func() {
		committed = true
	})

	sub := event.NewSubscription(100, []event.Event_Topic{event.Event_ContractUserEvent, event.Event_ContractSystemEvent})
	ec := event.GetEventCollectorInstance()
	ec.Subscribe(sub)
	defer ec.Unsubscribe(sub)

	count0 := 0
	count1 := 0
	go func() {
		for {
			select {
			case e := <-sub.ReadChan():
				t.Log(e.String())
				if e.Topic == event.Event_ContractUserEvent {
					count0++
				} else if e.Topic == event.Event_ContractSystemEvent {
					count1++
				}
			}
		}
	}()

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
	time.Sleep(10 * time.Millisecond)

	if count0 != 1 || count1 != 1 {
		t.Fatalf("expect count0 = 1, count1 = 1, got %d, %d", count0, count1)
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
		Publisher:  &crypto.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "Contracttestjs",
		ActionName: "hello",
		Data:       `[]`,
	}

	mtx.Actions = append(mtx.Actions, &ac)

	c := contract.Contract{
		ID: "Contracttestjs",
		Code: `
class Contract {
 constructor() {
  
 }
 hello() {
	console.log("hello");
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

	db.EXPECT().Get("state", "c-Contracttestjs").DoAndReturn(func(table string, key string) (string, error) {
		return c.Encode(), nil
	})

	db.EXPECT().Get("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000000)), nil
	})

	db.EXPECT().Get("state", "i-witness-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ-b", gomock.Any()).Times(2).DoAndReturn(func(table string, key string, value string) error {

		if database.MustUnmarshal(value).(int64) != 999993 {
			t.Fatal(database.MustUnmarshal(value).(int64))
		}

		return nil
	})

	db.EXPECT().Put("state", "i-witness-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
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
