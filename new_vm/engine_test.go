package new_vm

import (
	"testing"

	"time"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	blk "github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
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
	pm.vms["javascript"] = vm

	//nvm := native_vm.VM{}
	//
	//pm.vms["native"] = &nvm
	staticMonitor = pm
	return bh, db, vm
}

func TestNewEngine(t *testing.T) { // test of normal engine work
	bh, db, vm := engineinit(t)
	e := NewEngine(bh, db)
	blkInfo := string(e.(*EngineImpl).host.ctx.Value("block_info").(database.SerializedJSON))
	if blkInfo != `{"number":"10","parent_hash":"abc","time":"123456","witness":"witness"}` {
		t.Fatal(blkInfo)
	}

	ac, err := account.NewAccount(nil)
	if err != nil {
		t.Fatal(err)
	}

	mtx := tx.Tx{
		Id:         "txid",
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: ac.Pubkey},
	}

	act := tx.Action{
		Contract:   "contract",
		ActionName: "abi",
		Data:       "datas",
	}

	mtx.Actions = append(mtx.Actions, act)

	c := contract.Contract{
		ID:   "contract",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "javascript",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				&contract.ABI{
					Name:     "abi",
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

	vm.EXPECT().LoadAndCall(gomock.Any(), gomock.Any(), "abi", "datas").DoAndReturn(
		func(host *Host, contract *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			return nil, nil, nil
		},
	)

	e.Exec(&mtx)

}

func TestCost(t *testing.T) { // tests of context transport
	bh, db, vm := engineinit(t)
	e := NewEngine(bh, db)
	blkInfo := string(e.(*EngineImpl).host.ctx.Value("block_info").(database.SerializedJSON))
	if blkInfo != `{"number":"10","parent_hash":"abc","time":"123456","witness":"witness"}` {
		t.Fatal(blkInfo)
	}

	mtx := tx.Tx{
		Id:         "txid",
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

	ac := tx.Action{
		Contract:   "contract",
		ActionName: "abi",
		Data:       "datas",
	}

	mtx.Actions = append(mtx.Actions, ac)

	ac2 := tx.Action{
		Contract:   "contract",
		ActionName: "abi",
		Data:       "data2",
	}
	mtx.Actions = append(mtx.Actions, ac2)

	c := contract.Contract{
		ID:   "contract",
		Code: "codes",
		Info: &contract.Info{
			Lang:        "javascript",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				&contract.ABI{
					Name:     "abi",
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
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-g-contract").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-witness").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Put("state", "i-IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value) != int64(900) {
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

	db.EXPECT().Put("state", "i-g-contract", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value) != int64(900) {
			t.Fatal(database.MustUnmarshal(value))
		}
		return nil
	})

	vm.EXPECT().LoadAndCall(gomock.Any(), gomock.Any(), "abi", "datas").DoAndReturn(
		func(host *Host, c *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.NewCost(0, 0, 100)
			//fmt.Println("call datas")
			return nil, cost, nil
		},
	)

	vm.EXPECT().LoadAndCall(gomock.Any(), gomock.Any(), "abi", "data2").DoAndReturn(
		func(host *Host, c *contract.Contract, api string, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			host.ABIConfig("payment", "contract_pay")
			cost = contract.NewCost(0, 0, 100)
			//fmt.Println("call data2")
			return nil, cost, nil
		},
	)

	e.Exec(&mtx)
}

func TestNative_Transfer(t *testing.T) { // tests of native vm works
	bh, db, _ := engineinit(t)
	e := NewEngine(bh, db)
	blkInfo := string(e.(*EngineImpl).host.ctx.Value("block_info").(database.SerializedJSON))
	if blkInfo != `{"number":"10","parent_hash":"abc","time":"123456","witness":"witness"}` {
		t.Fatal(blkInfo)
	}

	mtx := tx.Tx{
		Id:         "txid",
		Time:       time.Now().UnixNano(),
		Expiration: 10000,
		GasLimit:   100000,
		GasPrice:   1,
		Publisher:  common.Signature{Pubkey: account.GetPubkeyByID("IOST8k3qxCkt4HNLGqmVdtxN7N1AnCdodvmb9yX4tUWzRzwWEx7sbQ")},
	}

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
				&contract.ABI{
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

	db.EXPECT().Put("state", "i-a", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(900) {
			t.Fatal("a", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1100) {
			t.Fatal("a", database.MustUnmarshal(value).(int64))
		}
		return nil
	})
	e.Exec(&mtx)

}
func TestNative_TopUp(t *testing.T) { // tests of native vm works
	bh, db, _ := engineinit(t)
	e := NewEngine(bh, db)
	blkInfo := string(e.(*EngineImpl).host.ctx.Value("block_info").(database.SerializedJSON))
	if blkInfo != `{"number":"10","parent_hash":"abc","time":"123456","witness":"witness"}` {
		t.Fatal(blkInfo)
	}

	mtx := tx.Tx{
		Id:         "txid",
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
				&contract.ABI{
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

	db.EXPECT().Get("state", "i-g-a").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Get("state", "i-b").DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	db.EXPECT().Put("state", "i-g-a", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(1100) {
			t.Fatal("g-a", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	db.EXPECT().Put("state", "i-b", gomock.Any()).DoAndReturn(func(table string, key string, value string) error {
		if database.MustUnmarshal(value).(int64) != int64(900) {
			t.Fatal("a", database.MustUnmarshal(value).(int64))
		}
		return nil
	})

	e.Exec(&mtx)

}
