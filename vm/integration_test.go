package vm

import (
	"testing"

	"fmt"

	"os"

	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
	"github.com/iost-official/Go-IOS-Protocol/vm/native"
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

var systemContract = native.ABI()

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

	os.RemoveAll("mvcc")
	//mvccdb := replaceDB(t)

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

	e := newEngine(bh, vi)

	//e.SetUp("js_path", jsPath)
	e.SetUp("log_level", "debug")
	e.SetUp("log_enable", "")
	return e, vi
}

func MakeTx(act tx.Action) (*tx.Tx, error) {
	trx := tx.NewTx([]*tx.Action{&act}, nil, int64(100000), int64(1), int64(10000000))

	ac, err := account.NewAccount(common.Base58Decode(testID[1]))
	if err != nil {
		return nil, err
	}
	trx, err = tx.SignTx(trx, ac)
	if err != nil {
		return nil, err
	}
	return trx, nil
}

func TestIntergration_Transfer(t *testing.T) {

	e, vi := ininit(t)

	act := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], "100"))

	trx := tx.NewTx([]*tx.Action{&act}, nil, int64(10000), int64(1), int64(10000000))

	ac, err := account.NewAccount(common.Base58Decode(testID[1]))
	if err != nil {
		t.Fatal(err)
	}
	trx, err = tx.SignTx(trx, ac)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("trasfer succes case:")
	t.Log(e.Exec(trx))
	t.Log("balance of sender :", vi.Balance(testID[0]))
	t.Log("balance of receiver :", vi.Balance(testID[2]))

	act2 := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], "999896"))
	trx2 := tx.NewTx([]*tx.Action{&act2}, nil, int64(10000), int64(1), int64(10000000))
	trx2, err = tx.SignTx(trx2, ac)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("trasfer not enough balance case:")
	t.Log(e.Exec(trx2))
	t.Log("balance of sender :", vi.Balance(testID[0]))
	t.Log("balance of receiver :", vi.Balance(testID[2]))
}

func jsHelloWorld() *contract.Contract {
	jshw := contract.Contract{
		ID: "ContractjsHelloWorld",
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

	trx, err := MakeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	ilog.Debugf(fmt.Sprintln(e.Exec(trx)))
	ilog.Debugf(fmt.Sprintln("balance of sender :", vi.Balance(testID[0])))

	act2 := tx.NewAction("Contract"+common.Base58Encode(trx.Hash()), "hello", `[]`)

	trx2, err := MakeTx(act2)
	if err != nil {
		t.Fatal(err)
	}

	ilog.Debugf(fmt.Sprintln(e.Exec(trx2)))
	ilog.Debugf(fmt.Sprintln("balance of sender :", vi.Balance(testID[0])))
}

func TestEngine_InitSetCode(t *testing.T) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll("mvcc")

	vi := database.NewVisitor(0, mvccdb)
	vi.SetBalance(testID[0], 1000000)
	vi.SetContract(systemContract)
	vi.Commit()

	bh := &block.BlockHead{
		ParentHash: []byte("abc"),
		Number:     0,
		Witness:    "witness",
		Time:       123456,
	}

	e := newEngine(bh, vi)

	//e.SetUp("js_path", jsPath)
	e.SetUp("log_level", "debug")
	e.SetUp("log_enable", "")

	jshw := jsHelloWorld()

	act := tx.NewAction("iost.system", "InitSetCode", fmt.Sprintf(`["iost.test", "%v"]`, jshw.B64Encode()))

	trx, err := MakeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	r, err := e.Exec(trx)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status.Code != tx.Success {
		t.Fatal(r)
	}
	ilog.Debugf(fmt.Sprintln("balance of sender :", vi.Balance(testID[0])))

	act2 := tx.NewAction("iost.test", "hello", `[]`)

	trx2, err := MakeTx(act2)
	if err != nil {
		t.Fatal(err)
	}

	r, err = e.Exec(trx2)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status.Code != tx.Success {
		t.Fatal(r)
	}
	ilog.Debugf(fmt.Sprintln("balance of sender :", vi.Balance(testID[0])))
}

func TestIntergration_CallJSCode(t *testing.T) {
	e, vi := ininit(t)

	jshw := jsHelloWorld()
	jsc := jsCallHelloWorld()

	vi.SetContract(jshw)
	vi.SetContract(jsc)

	act := tx.NewAction("Contractcall_hello_world", "call_hello", fmt.Sprintf(`[]`))

	trx, err := MakeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(e.Exec(trx))
	t.Log("balance of sender :", vi.Balance(testID[0]))
}

func jsCallHelloWorld() *contract.Contract {
	return &contract.Contract{
		ID: "Contractcall_hello_world",
		Code: `
class Contract {
 constructor() {

 }
 call_hello() {
  return BlockChain.call("ContractjsHelloWorld", "hello", "[]")
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

func TestIntergration_CallJSCodeWithReceipt(t *testing.T) {
	e, vi := ininit(t)

	jshw := jsHelloWorld()
	jsc := jsCallHelloWorldWithReceipt()

	vi.SetContract(jshw)
	vi.SetContract(jsc)

	act := tx.NewAction("Contractcall_hello_world", "call_hello", fmt.Sprintf(`[]`))

	trx, err := MakeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(e.Exec(trx))
	t.Log("balance of sender :", vi.Balance(testID[0]))
}

func jsCallHelloWorldWithReceipt() *contract.Contract {
	return &contract.Contract{
		ID: "Contractcall_hello_world",
		Code: `
class Contract {
 constructor() {
  
 }
 call_hello() {
  return BlockChain.callWithReceipt("ContractjsHelloWorld", "hello", "[]")
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

	ilog.Debugf("init %v", jshw.Info.Abis[0].GetLimit())

	e, vi := ininit(t)
	vi.SetContract(jshw)

	vi.SetBalance("CGjsHelloWorld", 1000000)

	act := tx.NewAction("ContractjsHelloWorld", "hello", fmt.Sprintf(`[]`))

	trx, err := MakeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	r, err := e.Exec(trx)
	ilog.Debugf("success: %v, %v", r, err)
	ilog.Debugf("balance of sender : %v", vi.Balance(testID[0]))
	ilog.Debugf("balance of contract : %v", vi.Balance("CGjsHelloWorld"))

}

func TestIntergration_Payment_Failed(t *testing.T) {
	jshw := jsHelloWorld()
	jshw.Info.Abis[0].Payment = 1
	jshw.Info.Abis[0].GasPrice = int64(10)

	jshw.Info.Abis[0].Limit.Data = -1
	jshw.Info.Abis[0].Limit.CPU = -1
	jshw.Info.Abis[0].Limit.Net = -1

	ilog.Debugf("init %v", jshw.Info.Abis[0].GetLimit())

	e, vi := ininit(t)
	vi.SetContract(jshw)

	vi.SetBalance("CGjsHelloWorld", 1000000)
	vi.Commit()

	act := tx.NewAction("ContractjsHelloWorld", "hello", fmt.Sprintf(`[]`))

	trx, err := MakeTx(act)
	if err != nil {
		t.Fatal(err)
	}

	r, err := e.Exec(trx)
	ilog.Debugf("success: %v, %v", r, err)
	ilog.Debugf("balance of sender : %v", vi.Balance(testID[0]))
	ilog.Debugf("balance of contract : %v", vi.Balance("CGjsHelloWorld"))

}

type fataler interface {
	Fatal(args ...interface{})
}

type JSTester struct {
	t      fataler
	e      Engine
	vi     *database.Visitor
	mvccdb db.MVCCDB

	cname string
	c     *contract.Contract
}

func NewJSTester(t fataler) *JSTester {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		panic(err)
	}

	os.RemoveAll("mvcc")

	//mvccdb := replaceDB(t)

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

	e := newEngine(bh, vi)

	e.SetUp("js_path", jsPath)
	//e.SetUp("log_level", "debug")
	//e.SetUp("log_enable", "")
	return &JSTester{
		t:      t,
		vi:     vi,
		e:      e,
		mvccdb: mvccdb,
	}
}

func (j *JSTester) ReadDB(key string) (value interface{}) {
	return database.MustUnmarshal(j.vi.Get(j.cname + "-" + key))
}

func (j *JSTester) FlushDB(t *testing.T, keys []string) {
	for _, k := range keys {
		t.Logf("%s: %v", k, j.ReadDB(k))
	}
}

func (j *JSTester) SetJS(code string) {
	j.c = &contract.Contract{
		ID:   "jsContract",
		Code: code,
		Info: &contract.Info{
			Lang:        "javascript",
			VersionCode: "1.0.0",
			Abis: []*contract.ABI{
				{
					Name:     "constructor",
					Args:     []string{},
					Payment:  0,
					GasPrice: int64(1),
					Limit:    contract.NewCost(100, 100, 100),
				},
			},
		},
	}
}

func (j *JSTester) DoSet() *tx.TxReceipt {
	act := tx.NewAction("iost.system", "SetCode", fmt.Sprintf(`["%v"]`, j.c.B64Encode()))

	trx, err := MakeTx(act)
	if err != nil {
		j.t.Fatal(err)
	}
	r, err := j.e.Exec(trx)
	if err != nil {
		j.t.Fatal(err)
	}
	j.cname = "Contract" + common.Base58Encode(trx.Hash())

	return r
}

func (j *JSTester) SetAPI(name string, argType ...string) {

	j.c.Info.Abis = append(j.c.Info.Abis, &contract.ABI{
		Name:     name,
		Payment:  0,
		GasPrice: int64(1),
		Limit:    contract.NewCost(100, 100, 100),
		Args:     argType,
	})

}

func (j *JSTester) TestJS(main, args string) *tx.TxReceipt {

	act2 := tx.NewAction(j.cname, main, args)

	trx2, err := MakeTx(act2)
	if err != nil {
		j.t.Fatal(err)
	}

	r, err := j.e.Exec(trx2)
	if err != nil {
		j.t.Fatal(err)
	}
	return r
}

func (j *JSTester) Clear() {
	j.mvccdb.Close()
	os.RemoveAll("mvcc")
}

func TestJSAPI_Database(t *testing.T) {
	js := NewJSTester(t)

	js.SetJS(`
class Contract {
	constructor() {
	this.aa = new Int64(100);
	}
	main() {
		this.aa = new Int64(45);
	}
}

module.exports = Contract;
`)
	js.SetAPI("main")
	js.DoSet()

	r := js.TestJS("main", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of publisher :", js.vi.Balance(testID[0]))
	t.Log("balance of receiver :", js.vi.Balance(testID[2]))
	t.Log("value of this.aa :", js.ReadDB("aa"))
}

func TestJSAPI_Transfer(t *testing.T) {

	js := NewJSTester(t)
	js.SetJS(`
class Contract {
	constructor() {
	}
	main() {
		BlockChain.transfer("IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY", "100")
	}
}

module.exports = Contract;
`)
	js.SetAPI("main")
	js.DoSet()

	r := js.TestJS("main", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.Balance(testID[0]))
	t.Log("balance of receiver :", js.vi.Balance(testID[2]))
}

func TestJSAPI_Transfer_Failed(t *testing.T) {

	js := NewJSTester(t)
	js.SetJS(`
class Contract {
	constructor() {
	}
	main() {
		BlockChain.transfer("IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd", "IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY", "100")
	}
}

module.exports = Contract;
`)
	js.SetAPI("main")
	js.DoSet()

	r := js.TestJS("main", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.Balance(testID[0]))
	t.Log("balance of receiver :", js.vi.Balance(testID[2]))
}

func TestJSAPI_Transfer_WrongFormat1(t *testing.T) {

	js := NewJSTester(t)
	js.SetJS(`
class Contract {
	constructor() {
	}
	main() {
		var ret = BlockChain.transfer("a", "b", 1);
		if (ret !== 0) {
			throw new Error("ret = ", ret);
		}
	}
}

module.exports = Contract;
`)
	js.SetAPI("main")
	js.DoSet()

	r := js.TestJS("main", fmt.Sprintf(`[]`))
	//todo wrong receipt
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.Balance(testID[0]))
	t.Log("balance of receiver :", js.vi.Balance(testID[2]))
}

func TestJSAPI_Deposit(t *testing.T) {

	js := NewJSTester(t)
	js.SetJS(`
class Contract {
	constructor() {
	}
	deposit() {
		return BlockChain.deposit("IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "100")
	}
	withdraw() {
		return BlockChain.withdraw("IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "99")
	}
}

module.exports = Contract;
`)
	js.SetAPI("deposit")
	js.SetAPI("withdraw")
	js.DoSet()

	r := js.TestJS("deposit", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.Balance(testID[0]))
	if 100 != js.vi.Balance(host.ContractAccountPrefix+js.cname) {
		t.Fatalf("balance of contract " + js.cname + "should be 100.")
	}

	r = js.TestJS("withdraw", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.Balance(testID[0]))
	if 1 != js.vi.Balance(host.ContractAccountPrefix+js.cname) {
		t.Fatalf("balance of contract " + js.cname + "should be 1.")
	}
}

func TestJSAPI_Info(t *testing.T) {

	js := NewJSTester(t)
	js.SetJS(`
class Contract {
	constructor() {
	}
	blockInfo() {
		var info = BlockChain.blockInfo()
		var obj = JSON.parse(info)
		_native_log(obj["parent_hash"])
		return obj["parent_hash"]
	}
	txInfo() {
		var info = BlockChain.txInfo()
		var obj = JSON.parse(info)
		_native_log(obj["hash"])
		return obj["hash"]
	}
}

module.exports = Contract;
`)
	js.SetAPI("blockInfo")
	js.SetAPI("txInfo")
	js.DoSet()

	r := js.TestJS("blockInfo", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)

	r = js.TestJS("txInfo", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
}

func TestJSRequireAuth(t *testing.T) {

	js := NewJSTester(t)
	js.SetJS(`
class Contract {
	constructor() {
	}
	requireAuth() {
		var ok = BlockChain.requireAuth("haha")
		_native_log(JSON.stringify(ok))
		ok = BlockChain.requireAuth("IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn")
		_native_log(JSON.stringify(ok))
		return ok
	}
}

module.exports = Contract;
`)
	js.SetAPI("requireAuth")
	js.DoSet()

	r := js.TestJS("requireAuth", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
}

func TestJS_Database(t *testing.T) {
	js := NewJSTester(t)
	defer js.Clear()
	lc, err := ReadFile("test_data/database.js")
	if err != nil {
		t.Fatal(err)
	}
	js.SetJS(string(lc))
	js.SetAPI("read")
	js.SetAPI("change")
	js.DoSet()
	t.Log("========= constructor")
	t.Log("num is ", js.ReadDB("num"))
	t.Log("string is ", js.ReadDB("string"))
	t.Log("bool is ", js.ReadDB("bool"))
	t.Log("nil is ", js.ReadDB("nil"))
	t.Log("array is ", js.ReadDB("array"))
	t.Log("object is ", js.ReadDB("object"))
	t.Log("arrayobj is ", js.ReadDB("arrayobj"))
	t.Log("objobj is ", js.ReadDB("objobj"))
	t.Log("========= read")

	js.TestJS("read", `[]`)
	//t.Log("num is ", js.ReadDB("num"))
	//t.Log("string is ", js.ReadDB("string"))
	//t.Log("bool is ", js.ReadDB("bool"))
	//t.Log("array is ", js.ReadDB("array"))
	//t.Log("object is ", js.ReadDB("object"))
	//t.Log("arrayobj is ", js.ReadDB("arrayobj"))
	//t.Log("objobj is ", js.ReadDB("objobj"))
	js.TestJS("change", `[]`)
	t.Log("========= change")
	t.Log("array is ", js.ReadDB("array"))
	t.Log("object is ", js.ReadDB("object"))
	t.Log("arrayobj is ", js.ReadDB("arrayobj"))
	t.Log("objobj is ", js.ReadDB("objobj"))
	t.Log("keyobj is", js.ReadDB("key"))
}

func TestJS_LuckyBet(t *testing.T) {
	js := NewJSTester(t)
	defer js.Clear()
	lc, err := ReadFile("test_data/lucky_bet.js")
	if err != nil {
		t.Fatal(err)
	}
	js.SetJS(string(lc))
	js.SetAPI("clearUserValue")
	js.SetAPI("bet", "string", "number", "number")
	js.SetAPI("getReward")
	js.DoSet()

	// here put the first bet
	r := js.TestJS("bet", fmt.Sprintf(`["%v",0, 2]`, testID[0]))
	t.Log("receipt is ", r)
	t.Log("max user number ", js.ReadDB("maxUserNumber"))
	t.Log("user count ", js.ReadDB("userNumber"))
	t.Log("total coins ", js.ReadDB("totalCoins"))
	t.Log("table should be saved ", js.ReadDB("table0"))

	for i := 1; i < 3; i++ { // at i = 2, should get reward
		r = js.TestJS("bet", fmt.Sprintf(`["%v",%v, %v]`, testID[0], i, i%4+1))
		if r.Status.Code != 0 {
			t.Fatal(r)
		}
	}

	t.Log("user count ", js.ReadDB("userNumber"))
	t.Log("total coins ", js.ReadDB("totalCoins"))
	t.Log("tables", js.ReadDB("tables"))
	t.Log("result 0 is ", js.ReadDB("result0"))
	t.Log("round is ", js.ReadDB("round"))
	for i := 3; i < 6; i++ { // at i = 6, should get reward 2nd times
		r = js.TestJS("bet", fmt.Sprintf(`["%v",%v, %v]`, testID[0], i, i%4+1))
		if r.Status.Code != 0 {
			t.Fatal(r)
		}
	}
	t.Log("round is ", js.ReadDB("round"))

}

func TestJS_Vote1(t *testing.T) {
	js := NewJSTester(t)
	defer js.Clear()
	lc, err := ReadFile("test_data/vote.js")
	if err != nil {
		t.Fatal(err)
	}
	js.SetJS(string(lc))
	js.SetAPI("RegisterProducer", "string", "string", "string", "string")
	js.SetAPI("UpdateProducer", "string", "string", "string", "string")
	js.SetAPI("LogInProducer", "string")
	js.SetAPI("LogOutProducer", "string")
	js.SetAPI("UnregisterProducer", "string")
	js.SetAPI("Vote", "string", "string", "number")
	js.SetAPI("Unvote", "string", "string", "number")
	js.SetAPI("Stat")
	js.SetAPI("Init")
	for i := 0; i <= 18; i += 2 {
		js.vi.SetBalance(testID[i], 5e+7)
	}
	js.vi.Commit()
	t.Log(js.DoSet())
	t.Log(js.TestJS("Init", `[]`))
	for i := 6; i <= 18; i += 2 {
		t.Log(js.vi.Balance(testID[i]))
	}

	keys := []string{
		"producerRegisterFee", "producerNumber", "preProducerThreshold", "preProducerMap",
		"voteLockTime", "currentProducerList", "pendingProducerList", "pendingBlockNumber",
		"producerTable",
		"voteTable",
	}
	js.FlushDB(t, keys)

	t.Log(js.vi.Balance(testID[18]))
	t.Log(js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[0])))
	js.FlushDB(t, keys)

	// test require auth
	t.Log(js.vi.Balance(testID[18]))
	t.Log(js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[2])))
	js.FlushDB(t, keys)

	// get pending producer info
	t.Log(database.MustUnmarshal(js.vi.Get(js.cname + "-" + "pendingBlockNumber")))
	t.Log(reflect.TypeOf(database.MustUnmarshal(js.vi.Get(js.cname + "-" + "pendingBlockNumber"))))
	t.Log(database.MustUnmarshal(js.vi.Get(js.cname + "-" + "pendingProducerList")))
	t.Log(reflect.TypeOf(database.MustUnmarshal(js.vi.Get(js.cname + "-" + "pendingProducerList"))))

	// test re register
	t.Log(js.vi.Balance(testID[18]))
	t.Log(js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[0])))
	js.FlushDB(t, keys)
}

func TestJS_VoteServi(t *testing.T) {
	js := NewJSTester(t)
	defer js.Clear()
	lc, err := ReadFile("test_data/vote.js")
	if err != nil {
		t.Fatal(err)
	}
	js.SetJS(string(lc))
	js.SetAPI("RegisterProducer", "string", "string", "string", "string")
	js.SetAPI("UpdateProducer", "string", "string", "string", "string")
	js.SetAPI("LogInProducer", "string")
	js.SetAPI("LogOutProducer", "string")
	js.SetAPI("UnregisterProducer", "string")
	js.SetAPI("Vote", "string", "string", "number")
	js.SetAPI("Unvote", "string", "string", "number")
	js.SetAPI("Stat")
	js.SetAPI("Init")
	for i := 0; i <= 18; i += 2 {
		js.vi.SetBalance(testID[i], 5e+7)
	}
	js.vi.Commit()
	t.Log(js.DoSet())
	t.Log(js.TestJS("Init", `[]`))
	keys := []string{
		"producerRegisterFee", "producerNumber", "preProducerThreshold", "preProducerMap",
		"voteLockTime", "currentProducerList", "pendingProducerList", "pendingBlockNumber",
		"producerTable",
		"voteTable",
	}
	js.FlushDB(t, keys)
}

func TestJS_Vote(t *testing.T) {
	js := NewJSTester(t)
	defer js.Clear()
	lc, err := ReadFile("test_data/vote.js")
	if err != nil {
		t.Fatal(err)
	}
	js.SetJS(string(lc))
	js.SetAPI("RegisterProducer", "string", "string", "string", "string")
	js.SetAPI("UpdateProducer", "string", "string", "string", "string")
	js.SetAPI("LogInProducer", "string")
	js.SetAPI("LogOutProducer", "string")
	js.SetAPI("UnregisterProducer", "string")
	js.SetAPI("Vote", "string", "string", "number")
	js.SetAPI("Unvote", "string", "string", "number")
	js.SetAPI("Stat")
	js.SetAPI("Init")
	for i := 0; i <= 18; i += 2 {
		js.vi.SetBalance(testID[i], 5e+7)
	}
	js.vi.Commit()
	t.Log(js.DoSet())
	t.Log(js.TestJS("Init", `[]`))

	keys := []string{
		"producerRegisterFee", "producerNumber", "preProducerThreshold", "preProducerMap",
		"voteLockTime", "currentProducerList", "pendingProducerList", "pendingBlockNumber",
		"producerTable",
		"voteTable",
	}
	js.FlushDB(t, keys)

	// test register, login, logout
	t.Log(js.TestJS("LogOutProducer", `["a"]`))
	t.Log(js.TestJS("LogInProducer", fmt.Sprintf(`["%v"]`, testID[0])))
	t.Log(js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[0])))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("LogInProducer", fmt.Sprintf(`["%v"]`, testID[0])))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("UpdateProducer", fmt.Sprintf(`["%v", "%v", "%v", "%v"]`, testID[0], "nloc", "nurl", "nnetid")))
	js.FlushDB(t, keys)

	// stat, no changes
	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	// vote and unvote
	t.Log(js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 10000000)))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 10000000)))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[2], 10000000)))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[2], testID[0], 1)))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 10000000)))
	js.FlushDB(t, keys)

	// stat testID[0] become pending producer
	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	bh := &block.BlockHead{
		ParentHash: []byte("abc"),
		Number:     211,
		Witness:    "witness",
		Time:       123456,
	}
	e := newEngine(bh, js.vi)
	e.SetUp("js_path", jsPath)
	js.e = e

	// test unvote
	t.Log(js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 20000001)))
	js.FlushDB(t, keys)

	t.Log(database.MustUnmarshal(js.vi.Get("i-" + testID[0] + "-s")))
	t.Log(js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 1000000)))
	js.FlushDB(t, keys)

	t.Log(js.vi.Servi(testID[0]))
	t.Log(js.vi.TotalServi())
	// stat pending producers don't get score
	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	// seven
	for i := 2; i <= 14; i += 2 {
		js.vi.SetBalance(testID[i], 5e+7)
	}
	for i := 2; i <= 14; i += 2 {
		t.Log(js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[i])))
		t.Log(js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[i], testID[i], 30000000+i)))
	}
	js.FlushDB(t, keys)

	// stat, offline producers don't get score
	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	for i := 2; i <= 14; i += 2 {
		t.Log(js.TestJS("LogInProducer", fmt.Sprintf(`["%v"]`, testID[i])))
	}
	js.FlushDB(t, keys)

	// stat, 1 producer become pending
	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("LogOutProducer", fmt.Sprintf(`["%v"]`, testID[12])))

	// stat, offline producer doesn't become pending. offline and pending producer don't get score, other pre producers get score
	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("LogInProducer", fmt.Sprintf(`["%v"]`, testID[12])))

	// stat, offline producer doesn't become pending. offline and pending producer don't get score, other pre producers get score
	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	// testID[0] become pre producer from pending producer, score = 0
	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 10000000)))
	js.FlushDB(t, keys)
	t.Log(js.vi.Servi(testID[0]))
	t.Log(js.vi.TotalServi())

	// unregister
	t.Log(js.TestJS("UnregisterProducer", fmt.Sprintf(`["%v"]`, testID[0])))
	js.FlushDB(t, keys)

	// unvote after unregister
	t.Log(js.TestJS("Unvote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[0], 9000000)))
	js.FlushDB(t, keys)
	t.Log(js.vi.Servi(testID[0]))
	t.Log(js.vi.TotalServi())

	// re register, score = 0, vote = 0
	t.Log(js.TestJS("RegisterProducer", fmt.Sprintf(`["%v","loc","url","netid"]`, testID[0])))
	t.Log(js.TestJS("LogInProducer", fmt.Sprintf(`["%v"]`, testID[0])))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Vote", fmt.Sprintf(`["%v", "%v", %d]`, testID[0], testID[2], 21000001)))
	js.FlushDB(t, keys)

	t.Log(js.TestJS("Stat", `[]`))
	js.FlushDB(t, keys)

	// unregister pre producer
	t.Log(js.TestJS("UnregisterProducer", fmt.Sprintf(`["%v"]`, testID[0])))
	js.FlushDB(t, keys)

	// test bonus
	t.Log(js.vi.Servi(testID[0]))
	t.Log(js.vi.Balance(host.ContractAccountPrefix + "iost.bonus"))
	act2 := tx.NewAction("iost.bonus", "ClaimBonus", fmt.Sprintf(`["%v", %d]`, testID[0], 1))

	trx2, err := MakeTx(act2)
	if err != nil {
		t.Fatal(err)
	}

	r, err := js.e.Exec(trx2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)

	t.Log(js.vi.Servi(testID[0]))
	t.Log(js.vi.Balance(host.ContractAccountPrefix + "iost.bonus"))
	t.Log(js.vi.Balance(testID[0]))
	act2 = tx.NewAction("iost.bonus", "ClaimBonus", fmt.Sprintf(`["%v", %d]`, testID[0], 21099999))

	trx2, err = MakeTx(act2)
	if err != nil {
		t.Fatal(err)
	}

	r, err = js.e.Exec(trx2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)

	t.Log(js.vi.Servi(testID[0]))
	t.Log(js.vi.Balance(host.ContractAccountPrefix + "iost.bonus"))
	t.Log(js.vi.Balance(testID[0]))
}
