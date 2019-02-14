package transfer

import (
	"context"
	"fmt"
	"github.com/iost-official/go-iost/sdk"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/iost-official/go-iost/test/performance/call"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
)

func init() {
	transfer := newTransferHandler()
	call.Register("transfer", transfer)
	iostSDK.SetChainID(chainID)
}

const (
	cache          = "transfer.cache"
	sep            = ","
	chainID uint32 = 1024
)

var rootKey = "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"
var iostSDK = sdk.NewIOSTDevSDK()

type transferHandler struct {
	testID     string
	contractID string
}

func newTransferHandler() *transferHandler {
	ret := &transferHandler{}
	ret.readCache()
	return ret
}

func (t *transferHandler) readCache() {
	content, err := ioutil.ReadFile(cache)
	if err == nil {
		strs := strings.Split(string(content), sep)
		if len(strs) > 1 {
			t.testID, t.contractID = strs[0], strs[1]
		}
	}
}

func (t *transferHandler) writeCache() {
	err := ioutil.WriteFile(cache, []byte(t.testID+sep+t.contractID), os.ModePerm)
	if err != nil {
		fmt.Println("write cache error: ", err)
		panic(err)
	}
}

// Prepare ...
func (t *transferHandler) Prepare() error {
	acc, _ := account.NewKeyPair(common.Base58Decode(rootKey), crypto.Ed25519)
	codePath := os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/test/performance/handles/transfer/transfer.js"
	abiPath := codePath + ".abi"
	client := call.GetClient(0)
	iostSDK.SetServer(client.Addr())
	iostSDK.SetAccount("admin", acc)
	iostSDK.SetTxInfo(500000.0, 1.0, 90, 0, nil)
	iostSDK.SetCheckResult(true, 3, 10)
	testKp, err := account.NewKeyPair(nil, crypto.Ed25519)
	if err != nil {
		return err
	}
	testID := "i" + strconv.FormatInt(time.Now().Unix(), 10)
	k := testKp.ReadablePubkey()
	_, err = iostSDK.CreateNewAccount(testID, k, k, 1000000, 10000, 100000)
	if err != nil {
		return err
	}
	err = iostSDK.PledgeForGasAndRAM(1500000, 0)
	if err != nil {
		return err
	}
	iostSDK.SetAccount(testID, testKp)
	_, txHash, err := iostSDK.PublishContract(codePath, abiPath, "", false, "")
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(30) * time.Second)
	resp, err := client.GetTxReceiptByTxHash(context.Background(), &rpcpb.TxHashRequest{Hash: txHash})
	if err != nil {
		return err
	}
	if tx.StatusCode(resp.StatusCode) != tx.Success {
		return fmt.Errorf("publish contract fail " + (resp.String()))
	}

	t.testID = testID
	t.contractID = "Contract" + txHash
	t.writeCache()
	return nil
}

// Run ...
func (t *transferHandler) Run(i int) (interface{}, error) {
	action := tx.NewAction(t.contractID, "transfer", fmt.Sprintf(`["admin","%v",1]`, t.testID))
	acc, _ := account.NewKeyPair(common.Base58Decode(rootKey), crypto.Ed25519)
	trx := tx.NewTx([]*tx.Action{action}, []string{}, 6000000, 100, time.Now().Add(time.Second*time.Duration(10000)).UnixNano(), 0, chainID)
	trx.AmountLimit = []*contract.Amount{{Token: "*", Val: "unlimited"}}
	stx, err := tx.SignTx(trx, "admin", []*account.KeyPair{acc})

	if err != nil {
		return nil, fmt.Errorf("sign tx error: %v", err)
	}
	var txHash string
	txHash, err = call.SendTx(stx, i)
	if err != nil {
		return nil, err
	}
	return txHash, nil
}
