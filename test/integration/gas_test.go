package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/vm/native"
)

func toString(n int64) string {
	return strconv.FormatInt(n, 10)
}

func toIOSTFixed(n int64) *common.Fixed {
	return &common.Fixed{Value: n * native.IOSTRatio, Decimal: 8}
}

const initCoin int64 = 5000
const contractName = "gas.iost"

var initCoinFN = toIOSTFixed(initCoin)
var monitor = vm.NewMonitor()

func gasTestInit() (*native.Impl, *host.Host, *contract.Contract, string, db.MVCCDB) {
	var tmpDB db.MVCCDB
	tmpDB, err := db.NewMVCCDB("mvcc")
	visitor := database.NewVisitor(100, tmpDB)
	if err != nil {
		panic(err)
	}
	context := host.NewContext(nil)
	context.Set("gas_price", int64(1))
	context.GSet("gas_limit", int64(100000))

	h := host.NewHost(context, visitor, monitor, nil)

	acc0Bytes, err := json.Marshal(acc0.ToAccount())
	if err != nil {
		panic(err)
	}
	h.DB().MPut("auth.iost"+"-auth", acc0.ID, database.MustMarshal(string(acc0Bytes)))

	acc1Bytes, err := json.Marshal(acc1.ToAccount())
	if err != nil {
		panic(err)
	}
	h.DB().MPut("auth.iost"+"-auth", acc1.ID, database.MustMarshal(string(acc1Bytes)))

	h.Context().Set("number", int64(1))
	h.Context().Set("time", int64(1541576370*1e9))
	h.Context().Set("stack_height", 0)
	h.Context().Set("publisher", acc0.ID)

	tokenContract := native.TokenABI()
	_, err = h.SetCode(tokenContract, "")
	if err != nil {
		panic(err)
	}

	authList := make(map[string]int)
	h.Context().Set("auth_contract_list", authList)
	authList[acc0.KeyPair.ID] = 2
	h.Context().Set("auth_list", authList)

	code := &contract.Contract{
		ID: contractName,
	}

	e := &native.Impl{}
	e.Init()

	h.Context().Set("contract_name", "token.iost")
	h.Context().Set("abi_name", "abi")
	h.Context().GSet("receipts", []*tx.Receipt{})
	_, _, err = e.LoadAndCall(h, tokenContract, "create", "iost", acc0.ID, int64(initCoin), []byte("{}"))
	if err != nil {
		panic("create iost " + err.Error())
	}
	_, _, err = e.LoadAndCall(h, tokenContract, "issue", "iost", acc0.ID, fmt.Sprintf("%d", initCoin))
	if err != nil {
		panic("issue iost " + err.Error())
	}
	if initCoin*1e8 != visitor.TokenBalance("iost", acc0.ID) {
		panic("set initial coins failed " + strconv.FormatInt(visitor.TokenBalance("iost", acc0.ID), 10))
	}

	h.Context().Set("contract_name", contractName)
	h.Context().Set("amount_limit", []*contract.Amount{&contract.Amount{Token:"*", Val:"unlimited"}})

	return e, h, code, acc0.ID, tmpDB
}

func timePass(h *host.Host, seconds int64) {
	h.Context().Set("time", h.Context().Value("time").(int64)+seconds*1e9)
}

func TestGas_NoPledge(t *testing.T) {
	Convey("test an account who did not pledge has 0 gas", t, func() {
		_, h, _, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		gas := h.GasManager.PGas(testAcc)
		So(gas.Value, ShouldEqual, 0)
	})
}

func TestGas_PledgeAuth(t *testing.T) {
	Convey("test pledging requires auth", t, func() {
		e, h, code, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		pledgeAmount := toIOSTFixed(200)
		authList := make(map[string]int)
		h.Context().Set("auth_list", authList)
		_, _, err := e.LoadAndCall(h, code, "pledge", testAcc, testAcc, pledgeAmount.ToString())
		So(err, ShouldNotBeNil)
	})
}

func TestGas_NotEnoughMoney(t *testing.T) {
	Convey("test pledging with not enough money", t, func() {
		e, h, code, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		pledgeAmount := toIOSTFixed(20000)
		_, _, err := e.LoadAndCall(h, code, "pledge", testAcc, testAcc, pledgeAmount.ToString())
		So(err, ShouldNotBeNil)
	})
}

func TestGas_Pledge(t *testing.T) {
	Convey("test pledge", t, func() {
		e, h, code, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		pledgeAmount := toIOSTFixed(200)
		_, _, err := e.LoadAndCall(h, code, "pledge", testAcc, testAcc, pledgeAmount.ToString())
		So(err, ShouldBeNil)
		So(h.DB().TokenBalance("iost", testAcc), ShouldEqual, initCoinFN.Value-pledgeAmount.Value)
		So(h.DB().TokenBalance("iost", contractName), ShouldEqual, pledgeAmount.Value)
		Convey("After pledge, you will get some gas immediately", func() {
			gas := h.GasManager.PGas(testAcc)
			gasEstimated := pledgeAmount.Multiply(native.GasImmediateReward)
			So(gas.Equals(gasEstimated), ShouldBeTrue)
		})
		Convey("Then gas increases at a predefined rate", func() {
			delta := int64(5)
			timePass(h, delta)
			gas := h.GasManager.PGas(testAcc)
			gasEstimated := pledgeAmount.Multiply(native.GasImmediateReward).Add(pledgeAmount.Multiply(native.GasIncreaseRate).Times(delta))
			So(gas.Equals(gasEstimated), ShouldBeTrue)
		})
		Convey("Then gas will reach limit and not increase any longer", func() {
			delta := int64(2 * native.GasFulfillSeconds)
			timePass(h, delta)
			gas := h.GasManager.PGas(testAcc)
			gasEstimated := pledgeAmount.Multiply(native.GasLimit)
			fmt.Printf("gas %v es %v\n", gas, gasEstimated)
			So(gas.Equals(gasEstimated), ShouldBeTrue)
		})
	})
}

func TestGas_PledgeMore(t *testing.T) {
	Convey("test you can pledge more after first time pledge", t, func() {
		e, h, code, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		firstTimePledgeAmount := toIOSTFixed(200)
		_, _, err := e.LoadAndCall(h, code, "pledge", testAcc, testAcc, firstTimePledgeAmount.ToString())
		So(err, ShouldBeNil)
		delta1 := int64(5)
		timePass(h, delta1)
		gasBeforeSecondPledge := h.GasManager.PGas(testAcc)
		secondTimePledgeAmount := toIOSTFixed(300)
		_, _, err = e.LoadAndCall(h, code, "pledge", testAcc, testAcc, secondTimePledgeAmount.ToString())
		So(err, ShouldBeNil)
		delta2 := int64(10)
		timePass(h, delta2)
		gasAfterSecondPledge := h.GasManager.PGas(testAcc)
		gasEstimated := gasBeforeSecondPledge.Add(secondTimePledgeAmount.Multiply(native.GasImmediateReward).Add(
			secondTimePledgeAmount.Add(firstTimePledgeAmount).Multiply(native.GasIncreaseRate).Times(delta2)))
		So(gasAfterSecondPledge.Equals(gasEstimated), ShouldBeTrue)
		So(h.DB().TokenBalance("iost", testAcc), ShouldEqual, initCoinFN.Sub(firstTimePledgeAmount).Sub(secondTimePledgeAmount).Value)
		So(h.DB().TokenBalance("iost", contractName), ShouldEqual, firstTimePledgeAmount.Add(secondTimePledgeAmount).Value)
	})
}

func TestGas_UseGas(t *testing.T) {
	Convey("test using gas", t, func() {
		e, h, code, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		pledgeAmount := int64(200)
		_, _, err := e.LoadAndCall(h, code, "pledge", testAcc, testAcc, toString(pledgeAmount))
		So(err, ShouldBeNil)
		delta1 := int64(5)
		timePass(h, delta1)
		gasBeforeUse := h.GasManager.PGas(testAcc)
		gasCost := toIOSTFixed(100)
		err = h.GasManager.CostGas(testAcc, gasCost)
		So(err, ShouldBeNil)
		gasAfterUse := h.GasManager.PGas(testAcc)
		gasEstimated := gasBeforeUse.Sub(gasCost)
		So(gasAfterUse.Equals(gasEstimated), ShouldBeTrue)
	})
}

func TestGas_unpledge(t *testing.T) {
	Convey("test unpledge", t, func() {
		e, h, code, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		pledgeAmount := toIOSTFixed(200)
		_, _, err := e.LoadAndCall(h, code, "pledge", testAcc, testAcc, pledgeAmount.ToString())
		So(err, ShouldBeNil)
		delta1 := int64(10)
		timePass(h, delta1)
		unpledgeAmount := toIOSTFixed(190)
		balanceBeforeunpledge := h.DB().TokenBalance("iost", testAcc)
		_, _, err = e.LoadAndCall(h, code, "unpledge", testAcc, testAcc, unpledgeAmount.ToString())
		So(err, ShouldBeNil)
		So(h.DB().TokenBalance("iost", testAcc), ShouldEqual, balanceBeforeunpledge)
		So(h.DB().TokenBalance("iost", contractName), ShouldEqual, pledgeAmount.Sub(unpledgeAmount).Value)
		gas := h.GasManager.PGas(testAcc)
		Convey("After unpledging, the gas limit will decrease. If current gas is more than the new limit, it will be decrease.", func() {
			gasEstimated := pledgeAmount.Sub(unpledgeAmount).Multiply(native.GasLimit)
			So(gas.Equals(gasEstimated), ShouldBeTrue)
		})
		Convey("after 3 days, the frozen money is available", func() {
			timePass(h, native.UnpledgeFreezeSeconds)
			h.Context().Set("contract_name", "token.iost")
			rs, _, err := e.LoadAndCall(h, native.TokenABI(), "balanceOf", "iost", testAcc)
			So(err, ShouldBeNil)
			expected := initCoinFN.Sub(pledgeAmount).Add(unpledgeAmount)
			So(rs[0], ShouldEqual, expected.ToString())
			So(h.DB().TokenBalance("iost", testAcc), ShouldEqual, expected.Value)
		})
	})
}

func TestGas_unpledgeTooMuch(t *testing.T) {
	Convey("test unpledge too much: each account has a minimum pledge", t, func() {
		e, h, code, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		pledgeAmount := int64(200)
		_, _, err := e.LoadAndCall(h, code, "pledge", testAcc, testAcc, toString(pledgeAmount))
		So(err, ShouldBeNil)
		delta1 := int64(1)
		timePass(h, delta1)
		unpledgeAmount := (pledgeAmount - int64(native.GasMinPledgeOfUser.ToFloat())) + int64(1)
		_, _, err = e.LoadAndCall(h, code, "unpledge", testAcc, testAcc, toString(unpledgeAmount))
		So(err, ShouldNotBeNil)
	})
}

func TestGas_PledgeunpledgeForOther(t *testing.T) {
	Convey("test pledge for others", t, func() {
		e, h, code, testAcc, tmpDB := gasTestInit()
		defer func() {
			tmpDB.Close()
			os.RemoveAll("mvcc")
		}()
		otherAcc := acc1.ID
		pledgeAmount := toIOSTFixed(200)
		_, _, err := e.LoadAndCall(h, code, "pledge", testAcc, otherAcc, pledgeAmount.ToString())
		h.FlushCacheCost()
		h.ClearCosts()
		So(err, ShouldBeNil)
		So(h.DB().TokenBalance("iost", testAcc), ShouldEqual, initCoinFN.Value-pledgeAmount.Value)
		So(h.DB().TokenBalance("iost", contractName), ShouldEqual, pledgeAmount.Value)
		Convey("After pledge, you will get some gas immediately", func() {
			gas := h.GasManager.PGas(otherAcc)
			gasEstimated := pledgeAmount.Multiply(native.GasImmediateReward)
			So(gas.Equals(gasEstimated), ShouldBeTrue)
		})
		Convey("If one pledge for others, he will get no gas himself", func() {
			gas := h.GasManager.PGas(testAcc)
			So(gas.Value, ShouldBeZeroValue)
		})
		Convey("Test unpledge for others", func() {
			unpledgeAmount := toIOSTFixed(190)
			_, _, err = e.LoadAndCall(h, code, "unpledge", testAcc, otherAcc, unpledgeAmount.ToString())
			So(err, ShouldBeNil)
			timePass(h, native.UnpledgeFreezeSeconds)
			h.Context().Set("contract_name", "token.iost")
			rs, _, err := e.LoadAndCall(h, native.TokenABI(), "balanceOf", "iost", testAcc)
			So(err, ShouldBeNil)
			expected := initCoinFN.Sub(pledgeAmount).Add(unpledgeAmount)
			So(rs[0], ShouldEqual, expected.ToString())
		})
	})
}

func TestGas_TGas(t *testing.T) {
	s := verifier.NewSimulator()
	defer s.Clear()
	createAccountsWithResource(s)
	ca, err := s.Compile("auth.iost", "../../contract/account", "../../contract/account.js")
	if err != nil {
		panic(err)
	}
	s.SetContract(ca)
	s.SetContract(native.GasABI())
	acc := prepareAuth(t, s)
	err = createToken(t, s, acc)
	if err != nil {
		panic(err)
	}
	other, err := account.NewKeyPair(nil, crypto.Secp256k1)
	otherID := "lispc0"
	s.Visitor.MPut("vote_producer.iost-producerTable", acc.ID, "dummy")
	Convey("test tgas", t, func() {
		Convey("account referrer should got 30000 tgas", func() {
			r, err := s.Call("auth.iost", "SignUp", array2json([]interface{}{otherID, other.ID, other.ID}), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.TGas(acc.ID).ToString(), ShouldEqual, "30000")
			r, err = s.Call("gas.iost", "pledge", array2json([]interface{}{acc.ID, otherID, "199"}), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldBeEmpty)
		})
		Convey("tgas can be transferred", func() {
			r, err := s.Call("gas.iost", "transfer", array2json([]interface{}{acc.ID, otherID, "10000"}), acc.ID, acc.KeyPair)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldEqual, "")
			So(s.Visitor.TGas(acc.ID).ToString(), ShouldEqual, "20000")
			So(s.Visitor.TGas(otherID).ToString(), ShouldEqual, "10000")
		})
		Convey("referrer get 15% reward", func() {
			s.Visitor.Commit()
			r, err := s.Call("token.iost", "transfer", array2json([]interface{}{"iost", otherID, acc.ID, "1", ""}), otherID, other)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldNotBeEmpty)
			So(s.Visitor.TGas(acc.ID).ToFloat(), ShouldAlmostEqual, 20000+float64(r.GasUsage)/100*0.15)
		})
		Convey("when pgas is used up, tgas will be used", func() {
			s.SetGas(otherID, 123)
			trx := tx.NewTx([]*tx.Action{{
				Contract:   "token.iost",
				ActionName: "transfer",
				Data:       array2json([]interface{}{"iost", otherID, acc.ID, "1", ""}),
			}}, nil, 1000000, 100, s.Head.Time+10000000, 0)
			trx.Time = s.Head.Time
			r, err := s.CallTx(trx, otherID, other)
			So(err, ShouldBeNil)
			So(r.Status.Message, ShouldNotBeEmpty)
			So(s.Visitor.TGas(otherID).ToFloat(), ShouldAlmostEqual, 10000-(r.GasUsage/100-123))
		})
	})

}
