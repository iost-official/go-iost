package luckyBet

import (
	"context"
	"fmt"
	"github.com/iost-official/go-iost/test/performance/call"
	"sync"

	"github.com/iost-official/go-iost/iwallet"

	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
	"google.golang.org/grpc"
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
	contractID string
}

func transfer(i int) {
	action := tx.NewAction(contractID, "bet", fmt.Sprintf("[\"%s\",%d,%d,%d]", testID, i%10, 1, 1))
	trx := tx.NewTx([]*tx.Action{action}, []string{}, 10000+int64(i), 100, time.Now().Add(time.Second*time.Duration(10000)).UnixNano(), 0)
	stx, err := tx.SignTx(trx, testID, []*account.KeyPair{testKp})
	if err != nil {
		fmt.Println("signtx", stx, err)
		return
	}
	var txHash string
	txHash, err = sendTx(stx, i)
	if err != nil {
		fmt.Println("sendtx", txHash, err)
		return
	}
}

func publish() string {
	codePath := "vm/test_data/lucky_bet.js"
	abiPath := codePath + ".abi"
	sdk := iwallet.SDK{}
	sdk.SetAccount(testID, testKp)
	sdk.SetTxInfo(10000.0, 1.0, 5, 0)
	_, txHash, err := sdk.PublishContract(codePath, abiPath, "", false, "")
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Duration(5) * time.Second)
	client := rpcpb.NewApiServiceClient(conns[0])
	resp, err := client.GetTxReceiptByTxHash(context.Background(), &rpcpb.TxHashRequest{Hash: txHash})
	if err != nil {
		panic(err)
	}
	if tx.StatusCode(resp.StatusCode) != tx.Success {
		panic("publish contract fail " + (resp.String()))
	}
	return "Contract" + txHash
}

func initAcc() {
	adminKp, err := account.NewKeyPair(loadBytes(rootKey), crypto.Ed25519)
	if err != nil {
		panic(err)
	}
	testKp, err = account.NewKeyPair(nil, crypto.Ed25519)
	if err != nil {
		panic(err)
	}
	testID = "testID"
	sdk := iwallet.SDK{}
	sdk.SetAccount("admin", adminKp)
	sdk.CreateNewAccount(testID, testKp, 100000, 10000, 100000)
}

func sendTx(stx *tx.Tx, i int) (string, error) {
	client := call.GetClient(i)
	resp, err := client.SendTransaction(context.Background(), call.ToTxRequest(stx))
	if err != nil {
		return "", err
	}
	return resp.Hash, nil
}

func loadBytes(s string) []byte {
	if s[len(s)-1] == 10 {
		s = s[:len(s)-1]
	}
	buf := common.Base58Decode(s)
	return buf
}
