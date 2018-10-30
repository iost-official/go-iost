package native

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
)

const initCoin int64 = 5000
const initNumber int64 = 10

func gasTestInit() (*host.Host, *account.Account) {
	var tmpDB db.MVCCDB
	tmpDB, err := db.NewMVCCDB(fmt.Sprintf("/tmp/gas_test%06d", rand.Intn(1000000)))
	visitor := database.NewVisitor(100, tmpDB)
	if err != nil {
		panic(err)
	}
	context := host.NewContext(nil)
	h := host.NewHost(context, visitor, nil, nil)
	testAcc := getTestAccount()
	as, err := json.Marshal(testAcc)
	if err != nil {
		panic(err)
	}
	h.DB().SetBalance(testAcc.ID, initCoin*IOSTRatio)
	h.DB().MPut("iost.auth-account", testAcc.ID, database.MustMarshal(string(as)))
	h.Context().Set("number", initNumber)
	authList := make(map[string]int)
	authList[testAcc.ID] = 2
	h.Context().Set("auth_list", authList)
	return h, testAcc
}

func getAccount(k string) *account.Account {
	key, err := account.NewKeyPair(common.Base58Decode(k), crypto.Ed25519)
	if err != nil {
		panic(err)
	}
	a := account.NewInitAccount(key.ID, key.ID, key.ID)
	return a
}

func getTestAccount() *account.Account {
	return getAccount("4nXuDJdU9MfP1TBY1W75o6ePDZNFuQ563YdkqVeEjW92aBcE6QDtFKPFWRBeKP8uMZcP7MGjfGubCLtu75t4ntxD")
}

func getGasAccount() *account.Account {
	return getAccount("2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1")
}

func getRootAccount() *account.Account {
	return getAccount("1rANSfcRzr4HkhbUFZ7L1Zp69JZZHiDDq5v7dNSbbEqeU4jxy3fszV4HGiaLQEyqVpS1dKT9g7zCVRxBVzuiUzB")
}

func min(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func TestGas_NoPledge(t *testing.T) {
	ilog.Info("test an account who did not pledge has 0 gas")
	h, testAcc := gasTestInit()
	gas := h.GasManager.CurrentGas(testAcc.ID)
	if gas != 0 {
		t.Fatalf("initial gas error %d", gas)
	}
}

func TestGas_PledgeAuth(t *testing.T) {
	ilog.Info("test pledging requires auth")
	h, testAcc := gasTestInit()
	pledgeAmount := int64(200)
	authList := make(map[string]int)
	h.Context().Set("auth_list", authList)
	_, _, err := pledgeGas.do(h, testAcc.ID, pledgeAmount)
	if err == nil {
		t.Fatalf("check auth should not succeed")
	}
}

func TestGas_Pledge(t *testing.T) {
	ilog.Info("test pledge")
	h, testAcc := gasTestInit()
	pledgeAmount := int64(200)
	_, _, err := pledgeGas.do(h, testAcc.ID, pledgeAmount)
	if err != nil {
		t.Fatalf("pledge err %v", err)
	}
	if h.DB().Balance(testAcc.ID) != (initCoin-pledgeAmount)*IOSTRatio {
		t.Fatalf("invalid balance after pledge %d", h.DB().Balance(testAcc.ID))
	}
	if h.DB().Balance(getGasAccount().ID) != pledgeAmount*IOSTRatio {
		t.Fatalf("invalid balance after pledge %d", h.DB().Balance(testAcc.ID))
	}
	ilog.Info("After pledge, you will get some gas immediately")
	gas := h.GasManager.CurrentGas(testAcc.ID)
	gasEstimated := pledgeAmount * host.GasImmediateReward
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
	ilog.Info("Then gas increases at a predefined rate")
	delta := int64(5)
	h.Context().Set("number", initNumber+delta)
	gas = h.GasManager.CurrentGas(testAcc.ID)
	gasEstimated = pledgeAmount*host.GasImmediateReward + delta*pledgeAmount*host.GasIncreaseRate
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
	ilog.Info("Then gas will reach limit and not increase any longer")
	delta = int64(100)
	h.Context().Set("number", initNumber+delta)
	gas = h.GasManager.CurrentGas(testAcc.ID)
	gasEstimated = pledgeAmount * host.GasLimit
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
}

func TestGas_PledgeMore(t *testing.T) {
	ilog.Info("test you can pledge more after first time pledge")
	h, testAcc := gasTestInit()
	firstTimePledgeAmount := int64(200)
	_, _, err := pledgeGas.do(h, testAcc.ID, firstTimePledgeAmount)
	if err != nil {
		t.Fatalf("pledge err %v", err)
	}
	delta1 := int64(5)
	h.Context().Set("number", initNumber+delta1)
	gasBeforeSecondPledge := h.GasManager.CurrentGas(testAcc.ID)
	secondTimePledgeAmount := int64(300)
	_, _, err = pledgeGas.do(h, testAcc.ID, secondTimePledgeAmount)
	if err != nil {
		t.Fatalf("pledge err %v", err)
	}
	delta2 := int64(10)
	h.Context().Set("number", initNumber+delta1+delta2)
	gasAfterSecondPledge := h.GasManager.CurrentGas(testAcc.ID)
	gasEstimated := gasBeforeSecondPledge + secondTimePledgeAmount*host.GasImmediateReward + (secondTimePledgeAmount+firstTimePledgeAmount)*host.GasIncreaseRate*delta2
	if gasAfterSecondPledge != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gasAfterSecondPledge, gasEstimated)
	}
	if h.DB().Balance(testAcc.ID) != (initCoin-firstTimePledgeAmount-secondTimePledgeAmount)*IOSTRatio {
		t.Fatalf("invalid balance after pledge %d", h.DB().Balance(testAcc.ID))
	}
	if h.DB().Balance(getGasAccount().ID) != (firstTimePledgeAmount+secondTimePledgeAmount)*IOSTRatio {
		t.Fatalf("invalid balance after pledge %d", h.DB().Balance(getGasAccount().ID))
	}
}

func TestGas_UseGas(t *testing.T) {
	ilog.Info("test using gas")
	h, testAcc := gasTestInit()
	pledgeAmount := int64(200)
	_, _, err := pledgeGas.do(h, testAcc.ID, pledgeAmount)
	if err != nil {
		t.Fatalf("pledge err %v", err)
	}
	delta1 := int64(5)
	h.Context().Set("number", initNumber+delta1)
	gasBeforeUse := h.GasManager.CurrentGas(testAcc.ID)
	gasCost := int64(100)
	err = h.GasManager.CostGas(testAcc.ID, gasCost)
	if err != nil {
		t.Fatalf("cost gas failed %v", err)
	}
	gasAfterUse := h.GasManager.CurrentGas(testAcc.ID)
	gasEstimated := gasBeforeUse - gasCost
	if gasAfterUse != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gasAfterUse, gasEstimated)
	}
}

func TestGas_Unpledge(t *testing.T) {
	ilog.Info("test unpledge")
	h, testAcc := gasTestInit()
	pledgeAmount := int64(200)
	_, _, err := pledgeGas.do(h, testAcc.ID, pledgeAmount)
	if err != nil {
		t.Fatalf("pledge err %v", err)
	}
	delta1 := int64(5)
	h.Context().Set("number", initNumber+delta1)
	unpledgeAmount := int64(100)
	_, _, err = unpledgeGas.do(h, testAcc.ID, unpledgeAmount)
	if err != nil {
		t.Fatalf("unpledge err %v", err)
	}
	if h.DB().Balance(testAcc.ID) != (initCoin-pledgeAmount+unpledgeAmount)*IOSTRatio {
		t.Fatalf("invalid balance after unpledge %d", h.DB().Balance(testAcc.ID))
	}
	if h.DB().Balance(getGasAccount().ID) != (pledgeAmount-unpledgeAmount)*IOSTRatio {
		t.Fatalf("invalid balance after unpledge %d", h.DB().Balance(testAcc.ID))
	}
	gas := h.GasManager.CurrentGas(testAcc.ID)
	ilog.Info("After unpledging, the gas limit will decrease. If current gas is more than the new limit, it will be decrease.")
	gasEstimated := (pledgeAmount - unpledgeAmount) * host.GasLimit
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
}
