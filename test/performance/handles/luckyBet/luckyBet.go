package luckyBet

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/iost-official/go-iost/iwallet"
	"github.com/iost-official/go-iost/test/performance/call"

	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
)

func init() {
	luckyBet := newLuckyBetHandler()
	call.Register("luckyBet", luckyBet)
}

const (
	cache = "luckyBet.cache"
	sep   = ","
)

var rootKey = "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"
var sdk = iwallet.SDK{}

type luckyBetHandler struct {
	testID     string
	testKp     *account.KeyPair
	contractID string
}

func newLuckyBetHandler() *luckyBetHandler {
	ret := &luckyBetHandler{}
	ret.readCache()
	return ret
}

func (t *luckyBetHandler) readCache() {
	content, err := ioutil.ReadFile(cache)
	if err == nil {
		strs := strings.Split(string(content), sep)
		if len(strs) > 2 {
			var secKey string
			t.testID, secKey, t.contractID = strs[0], strs[1], strs[2]
			t.testKp, err = account.NewKeyPair(common.Base58Decode(secKey), crypto.Ed25519)
			if err != nil {
				panic("readCache secKey error")
			}
		}
	}
}

func (t *luckyBetHandler) writeCache() {
	err := ioutil.WriteFile(cache, []byte(t.testID+sep+common.Base58Encode(t.testKp.Seckey)+sep+t.contractID), os.ModePerm)
	if err != nil {
		fmt.Println("write cache error: ", err)
		panic(err)
	}
}

// Prepare ...
func (t *luckyBetHandler) Prepare() error {
	log.Println("lucky bet Prepare")
	acc, _ := account.NewKeyPair(common.Base58Decode(rootKey), crypto.Ed25519)
	codePath := os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/vm/test_data/lucky_bet.js"
	abiPath := codePath + ".abi"
	client := call.GetClient(0)
	sdk.SetServer(client.Addr())
	sdk.SetAccount("admin", acc)
	sdk.SetTxInfo(3000000.0, 1.0, 90, 0)
	sdk.SetCheckResult(true, 3, 10)
	var err error
	t.testKp, err = account.NewKeyPair(nil, crypto.Ed25519)
	if err != nil {
		return err
	}
	testID := "i" + strconv.FormatInt(time.Now().Unix(), 10)
	err = sdk.CreateNewAccount(testID, t.testKp, 900000000, 100000000, 100000)
	if err != nil {
		return err
	}
	err = sdk.PledgeForGasAndRAM(15000000, 0)
	if err != nil {
		return err
	}
	sdk.SetAccount(testID, t.testKp)
	_, txHash, err := sdk.PublishContract(codePath, abiPath, "", false, "")
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
func (t *luckyBetHandler) Run(i int) (interface{}, error) {
	action := tx.NewAction(t.contractID, "bet", fmt.Sprintf(`["%v",%d,%d,%d]`, t.testID, i%10, 1, 1))
	trx := tx.NewTx([]*tx.Action{action}, []string{}, 10000000000+int64(i), 100, time.Now().Add(time.Second*time.Duration(10000)).UnixNano(), 0)
	stx, err := tx.SignTx(trx, t.testID, []*account.KeyPair{t.testKp})

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
