package native

import (
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	"testing"
)

func myInit() *host.Host {
	var tmpDB db.MVCCDB
	tmpDB, err := db.NewMVCCDB("/tmp/gas_test")
	visitor := database.NewVisitor(100, tmpDB)
	if err != nil {
		panic(err)
	}
	//monitor := vm.NewMonitor()
	context := host.NewContext(nil)
	host := host.NewHost(context, visitor, nil, nil)
	return host
}

func getAccount(k string) *account.Account {
	acc, err := account.NewAccount(common.Base58Decode(k), crypto.Ed25519)
	if err != nil {
		panic(err)
	}
	return acc

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

func TestGas_AllInOneTest(t *testing.T) {
	ilog.Info("Init the test environ")
	host := myInit()
	testAcc := getTestAccount()
	t.Logf("test with account %s", testAcc.ID)
	initCoin := int64(5000)
	host.DB().SetBalance(testAcc.ID, initCoin)
	initNumber := int64(10)
	host.Context().Set("number", initNumber)
	gas := host.GasManager.CurrentGas(testAcc.ID)
	if gas != 0 {
		t.Fatalf("initial gas error %d", gas)
	}
	ilog.Info("start test pledge")
	ilog.Info("test auth. It should failed now.")
	pledgeAmount := int64(200)
	authList := make(map[string]int)
	host.Context().Set("auth_list", authList)
	_, _, err := pledgeGas.do(host, testAcc.ID, pledgeAmount)
	if err == nil {
		t.Fatalf("check auth should not succeed")
	}
	ilog.Info("test pledge")
	authList[testAcc.ID] = 2
	host.Context().Set("auth_list", authList)
	_, _, err = pledgeGas.do(host, testAcc.ID, pledgeAmount)
	if err != nil {
		t.Fatalf("pledge err %v", err)
	}
	if host.DB().Balance(testAcc.ID) != (initCoin - pledgeAmount) {
		t.Fatalf("invalid balance after pledge %d", host.DB().Balance(testAcc.ID))
	}
	if host.DB().Balance(getGasAccount().ID) != pledgeAmount {
		t.Fatalf("invalid balance after pledge %d", host.DB().Balance(testAcc.ID))
	}
	ilog.Info("After pledge, check gas amount")
	delta := int64(0)
	gas = host.GasManager.CurrentGas(testAcc.ID)
	gasEstimated := min(pledgeAmount*gasLimitRatio, pledgeAmount*gasImmediateRatio+delta*pledgeAmount*gasRateRatio)
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
	ilog.Info("check gas increase rate")
	delta = int64(5)
	host.Context().Set("number", initNumber+delta)
	gas = host.GasManager.CurrentGas(testAcc.ID)
	gasEstimated = min(pledgeAmount*gasLimitRatio, pledgeAmount*gasImmediateRatio+delta*pledgeAmount*gasRateRatio)
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
	ilog.Info("check gas limit")
	delta = int64(100)
	host.Context().Set("number", initNumber+delta)
	gas = host.GasManager.CurrentGas(testAcc.ID)
	gasEstimated = min(pledgeAmount*gasLimitRatio, pledgeAmount*gasImmediateRatio+delta*pledgeAmount*gasRateRatio)
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
	ilog.Info("check gas amount after usage")
	gasCost := int64(100)
	err = host.GasManager.CostGas(testAcc.ID, gasCost)
	if err != nil {
		t.Fatalf("cost gas failed %v", err)
	}
	gas = host.GasManager.CurrentGas(testAcc.ID)
	gasEstimated = gasEstimated - gasCost
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
	ilog.Info("test unpledge")
	unpledgeAmount := int64(100)
	_, _, err = unpledgeGas.do(host, testAcc.ID, unpledgeAmount)
	if err != nil {
		t.Fatalf("unpledge err %v", err)
	}
	if host.DB().Balance(testAcc.ID) != (initCoin - pledgeAmount + unpledgeAmount) {
		t.Fatalf("invalid balance after unpledge %d", host.DB().Balance(testAcc.ID))
	}
	if host.DB().Balance(getGasAccount().ID) != (pledgeAmount - unpledgeAmount) {
		t.Fatalf("invalid balance after unpledge %d", host.DB().Balance(testAcc.ID))
	}
	gas = host.GasManager.CurrentGas(testAcc.ID)
	gasEstimated = (pledgeAmount - unpledgeAmount) * gasLimitRatio
	if gas != gasEstimated {
		t.Fatalf("invalid gas %d != %d", gas, gasEstimated)
	}
}
