package main

import (
	"context"
	"fmt"
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

var conns []*grpc.ClientConn
var rootKey = "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"
var contractID string

var testID string
var testKp *account.KeyPair

func initConn(num int) {
	conns = make([]*grpc.ClientConn, num)
	allServers := []string{"localhost:30002"}

	for i := 0; i < num; i++ {
		conn, err := grpc.Dial(allServers[i%len(allServers)], grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		conns[i] = conn
	}

}

func transParallel(num int) {
	wg := new(sync.WaitGroup)
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(i int) {
			transfer(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func toTxRequest(t *tx.Tx) *rpcpb.TransactionRequest {
	ret := &rpcpb.TransactionRequest{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasPrice:   float64(t.GasPrice) / 100,
		GasLimit:   float64(t.GasLimit) / 100,
		Delay:      t.Delay,
		Signers:    t.Signers,
		Publisher:  t.Publisher,
	}
	for _, a := range t.Actions {
		ret.Actions = append(ret.Actions, &rpcpb.Action{
			Contract:   a.Contract,
			ActionName: a.ActionName,
			Data:       a.Data,
		})
	}
	for _, a := range t.AmountLimit {
		fixed, err := common.UnmarshalFixed(a.Val)
		if err != nil {
			continue
		}
		ret.AmountLimit = append(ret.AmountLimit, &rpcpb.AmountLimit{
			Token: a.Token,
			Value: fixed.ToFloat(),
		})
	}
	for _, s := range t.Signs {
		ret.Signatures = append(ret.Signatures, &rpcpb.Signature{
			Algorithm: rpcpb.Signature_Algorithm(s.Algorithm),
			PublicKey: s.Pubkey,
			Signature: s.Sig,
		})
	}
	for _, s := range t.PublishSigns {
		ret.PublisherSigs = append(ret.PublisherSigs, &rpcpb.Signature{
			Algorithm: rpcpb.Signature_Algorithm(s.Algorithm),
			PublicKey: s.Pubkey,
			Signature: s.Sig,
		})
	}
	return ret
}

func sendTx(stx *tx.Tx, i int) (string, error) {
	client := rpcpb.NewApiServiceClient(conns[i])
	resp, err := client.SendTransaction(context.Background(), toTxRequest(stx))
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
	sdk.SetTxInfo(10000, 100, 5, 0)
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

func main() {

	var iterNum = 800
	var parallelNum = 30
	initConn(parallelNum)
	initAcc()

	contractID = publish()

	start := time.Now()

	for i := 0; i < iterNum; i++ {
		fmt.Println(i)
		transParallel(parallelNum)
	}

	fmt.Println("done. timecost=", time.Since(start))

}
