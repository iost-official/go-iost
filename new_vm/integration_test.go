package new_vm

import (
	"testing"

	"fmt"

	"os"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/native_vm"
)

var testID = []string{
	"IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE",
	"IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY", "8dJ9YKovJ5E7hkebAQaScaG1BA8snRUHPUbVcArcTVq6",
	"IOST7ZNDWeh8pHytAZdpgvp7vMpjZSSe5mUUKxDm6AXPsbdgDMAYhs", "7CnwT7BXkEFAVx6QZqC7gkDhQwbvC3d2CkMZvXHZdDMN",
	"IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd", "Htarc5Sp4trjqY4WrTLtZ85CF6qx87v7CRwtV4RRGnbF",
	"IOST7GmPn8xC1RESMRS6a62RmBcCdwKbKvk2ZpxZpcXdUPoJdapnnh", "Bk8bAyG4VLBcrsoRErPuQGhwCy4C1VxfKE4jjX9oLhv",
	"IOST7ZGQL4k85v4wAxWngmow7JcX4QFQ4mtLNjgvRrEnEuCkGSBEHN", "546aCDG9igGgZqVZeybajaorP5ZeF9ghLu2oLncXk3d6",
	"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9", "DXNYRwG7dRFkbWzMNEbKfBhuS8Yn51x9J6XuTdNwB11M",
	"IOST8mFxe4kq9XciDtURFZJ8E76B8UssBgRVFA5gZN9HF5kLUVZ1BB", "AG8uECmAwFis8uxTdWqcgGD9tGDwoP6CxqhkhpuCdSeC",
	"IOST7uqa5UQPVT9ongTv6KmqDYKdVYSx4DV2reui4nuC5mm5vBt3D9", "GJt5WSSv5WZi1axd3qkb1vLEfxCEgKGupcXf45b5tERU",
	"IOST6wYBsLZmzJv22FmHAYBBsTzmV1p1mtHQwkTK9AjCH9Tg5Le4i4", "7U3uwEeGc2TF3Xde2oT66eTx1Uw15qRqYuTnMd3NNjai",
}

var systemContract = native_vm.NativeABI()

func replaceDB(t *testing.T) database.IMultiValue {
	ctl := gomock.NewController(t)
	mvccdb := database.NewMockIMultiValue(ctl)

	var senderbalance int64
	var receiverbalance int64

	mvccdb.EXPECT().Get("state", "i-"+testID[0]).AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(senderbalance), nil
	})
	mvccdb.EXPECT().Get("state", "i-"+testID[2]).AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(receiverbalance), nil
	})

	mvccdb.EXPECT().Get("state", "c-iost.system").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return systemContract.Encode(), nil
	})

	mvccdb.EXPECT().Get("state", "i-witness").AnyTimes().DoAndReturn(func(table string, key string) (string, error) {
		return database.MustMarshal(int64(1000)), nil
	})

	mvccdb.EXPECT().Put("state", "c-iost.system", gomock.Any()).AnyTimes().DoAndReturn(func(table string, key string, value string) error {
		return nil
	})

	mvccdb.EXPECT().Put("state", "i-"+testID[0], gomock.Any()).AnyTimes().DoAndReturn(func(table string, key string, value string) error {
		t.Log("sender balance:", database.Unmarshal(value))
		senderbalance = database.Unmarshal(value).(int64)
		return nil
	})

	mvccdb.EXPECT().Put("state", "i-"+testID[2], gomock.Any()).AnyTimes().DoAndReturn(func(table string, key string, value string) error {
		t.Log("receiver balance:", database.Unmarshal(value))
		receiverbalance = database.Unmarshal(value).(int64)
		return nil
	})

	mvccdb.EXPECT().Put("state", "i-witness", gomock.Any()).AnyTimes().DoAndReturn(func(table string, key string, value string) error {

		//fmt.Println("witness received money", database.MustUnmarshal(value))
		//if database.MustUnmarshal(value) != int64(1100) {
		//	t.Fatal(database.MustUnmarshal(value))
		//}
		return nil
	})

	mvccdb.EXPECT().Rollback().Do(func() {
		t.Log("exec tx failed, and success rollback")
	})

	mvccdb.EXPECT().Commit().Do(func() {
		t.Log("committed")
	})

	return mvccdb
}

func ininit(t *testing.T) (Engine, *database.Visitor) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}

	//mvccdb := replaceDB(t)

	defer os.RemoveAll("mvcc")

	vi := database.NewVisitor(0, mvccdb)
	vi.SetBalance(testID[0], 1000000)
	vi.SetContract(systemContract)
	vi.Commit()

	bh := &block.BlockHead{
		ParentHash: []byte("abc"),
		Number:     10,
		Witness:    "witness",
		Time:       123456,
	}

	e := NewEngine(bh, mvccdb)

	e.SetUp("js_path", jsPath)
	e.SetUp("log_level", "debug")
	e.SetUp("log_enable", "")
	return e, vi
}

func makeTx(act tx.Action) (*tx.Tx, error) {
	trx := tx.NewTx([]tx.Action{act}, nil, int64(10000), int64(1), int64(10000000))

	ac, err := account.NewAccount(common.Base58Decode(testID[1]))
	if err != nil {
		return nil, err
	}
	trx, err = tx.SignTx(trx, ac)
	if err != nil {
		return nil, err
	}
	return &trx, nil
}

func TestIntergration_Transfer(t *testing.T) {

	e, vi := ininit(t)

	act := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], "100"))

	trx := tx.NewTx([]tx.Action{act}, nil, int64(10000), int64(1), int64(10000000))

	ac, err := account.NewAccount(common.Base58Decode(testID[1]))
	if err != nil {
		t.Fatal(err)
	}
	trx, err = tx.SignTx(trx, ac)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("trasfer succes case:")
	t.Log(e.Exec(&trx))
	t.Log("balance of sender :", vi.Balance(testID[0]))
	t.Log("balance of receiver :", vi.Balance(testID[2]))

	act2 := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], "999896"))
	trx2 := tx.NewTx([]tx.Action{act2}, nil, int64(10000), int64(1), int64(10000000))
	trx2, err = tx.SignTx(trx2, ac)

	t.Log("trasfer not enough balance case:")
	t.Log(e.Exec(&trx2))
	t.Log("balance of sender :", vi.Balance(testID[0]))
	t.Log("balance of receiver :", vi.Balance(testID[2]))
}

func jsHelloWorld() *contract.Contract {
	jshw := contract.Contract{
		ID: "jsHelloWorld",
		Code: `
class Contract {
 constructor() {
  
 }
 hello() {
  return "world";
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
				}, {
					Name:     "constructor",
					Payment:  0,
					GasPrice: int64(1),
					Limit:    contract.NewCost(100, 100, 100),
					Args:     []string{},
				},
			},
		},
	}
	return &jshw
}

func TestIntergration_SetCode(t *testing.T) {
	e, vi := ininit(t)

	jshw := jsHelloWorld()

	act := tx.NewAction("iost.system", "SetCode", fmt.Sprintf(`["%v"]`, jshw.B64Encode()))

	trx, err := makeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	ilog.Debug(fmt.Sprintln(e.Exec(trx)))
	ilog.Debug(fmt.Sprintln("balance of sender :", vi.Balance(testID[0])))

	act2 := tx.NewAction("Contract"+common.Base58Encode(trx.Hash()), "hello", `[]`)

	trx2, err := makeTx(act2)
	if err != nil {
		t.Fatal(err)
	}

	ilog.Debug(fmt.Sprintln(e.Exec(trx2)))
	ilog.Debug(fmt.Sprintln("balance of sender :", vi.Balance(testID[0])))
}

func TestIntergration_CallJSCode(t *testing.T) {
	e, vi := ininit(t)

	jshw := jsHelloWorld()
	jsc := jsCallHelloWorld()

	vi.SetContract(jshw)
	vi.SetContract(jsc)

	act := tx.NewAction("call_hello_world", "call_hello", fmt.Sprintf(`[]`))

	trx, err := makeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(e.Exec(trx))
	t.Log("balance of sender :", vi.Balance(testID[0]))
}

func jsCallHelloWorld() *contract.Contract {
	return &contract.Contract{
		ID: "call_hello_world",
		Code: `
class Contract {
 constructor() {
  
 }
 call_hello() {
  return BlockChain.call("jsHelloWorld", "hello", "[]")
 }
}

module.exports = Contract;
`,
		Info: &contract.Info{
			Lang:        "javascript",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "call_hello",
					Payment:  0,
					GasPrice: int64(1),
					Limit:    contract.NewCost(100, 100, 100),
					Args:     []string{},
				},
			},
		},
	}
}

func TestIntergration_Payment_Success(t *testing.T) {
	jshw := jsHelloWorld()
	jshw.Info.Abis[0].Payment = 1
	jshw.Info.Abis[0].GasPrice = int64(10)

	ilog.Debug("init %v", jshw.Info.Abis[0].GetLimit())

	e, vi := ininit(t)
	vi.SetContract(jshw)

	vi.SetBalance("CGjsHelloWorld", 1000000)

	act := tx.NewAction("jsHelloWorld", "hello", fmt.Sprintf(`[]`))

	trx, err := makeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	r, err := e.Exec(trx)
	ilog.Debug("success: %v, %v", r, err)
	ilog.Debug("balance of sender : %v", vi.Balance(testID[0]))
	ilog.Debug("balance of contract : %v", vi.Balance("CGjsHelloWorld"))

}

func TestIntergration_Payment_Failed(t *testing.T) {
	jshw := jsHelloWorld()
	jshw.Info.Abis[0].Payment = 1
	jshw.Info.Abis[0].GasPrice = int64(10)

	jshw.Info.Abis[0].Limit.Data = -1
	jshw.Info.Abis[0].Limit.CPU = -1
	jshw.Info.Abis[0].Limit.Net = -1

	ilog.Debug("init %v", jshw.Info.Abis[0].GetLimit())

	e, vi := ininit(t)
	vi.SetContract(jshw)

	vi.SetBalance("CGjsHelloWorld", 1000000)
	vi.Commit()

	act := tx.NewAction("jsHelloWorld", "hello", fmt.Sprintf(`[]`))

	trx, err := makeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	r, err := e.Exec(trx)
	ilog.Debug("success: %v, %v", r, err)
	ilog.Debug("balance of sender : %v", vi.Balance(testID[0]))
	ilog.Debug("balance of contract : %v", vi.Balance("CGjsHelloWorld"))

}
