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
	"github.com/iost-official/go-iost/rpc"
	"google.golang.org/grpc"
)

var conns []*grpc.ClientConn
var rootKey = "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"
var contractID string
var sdk = iwallet.SDK{}

func initConn(num int) {
	conns = make([]*grpc.ClientConn, num)
	allServers := []string{"18.182.146.155:30002", "54.64.205.127:30002"}
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

func sendTx(stx *tx.Tx, i int) ([]byte, error) {
	client := rpc.NewApisClient(conns[i])
	resp, err := client.SendTx(context.Background(), &rpc.TxReq{Tx: stx.ToPb()})
	if err != nil {
		return nil, err
	}
	return []byte(resp.Hash), nil
}

func loadBytes(s string) []byte {
	if s[len(s)-1] == 10 {
		s = s[:len(s)-1]
	}
	buf := common.Base58Decode(s)
	return buf
}

func transfer(i int) {
	action := tx.NewAction(contractID, "transfer", `["admin","testID",1]`)
	acc, _ := account.NewKeyPair(loadBytes(rootKey), crypto.Ed25519)
	trx := tx.NewTx([]*tx.Action{action}, []string{}, 1000, 100, time.Now().Add(time.Second*time.Duration(10000)).UnixNano(), 0)
	stx, err := tx.SignTx(trx, "admin", []*account.KeyPair{acc})

	if err != nil {
		fmt.Println("signtx", stx, err)
		return
	}
	var txHash []byte
	txHash, err = sendTx(stx, i)
	if err != nil {
		fmt.Println("sendtx", txHash, err)
		return
	}
}

func publish() string {
	acc, _ := account.NewKeyPair(loadBytes(rootKey), crypto.Ed25519)
	codePath := "transfer.js"
	abiPath := codePath + ".abi"
	sdk.SetAccount("admin", acc)
	sdk.SetServer("54.95.152.91:30002")
	sdk.SetTxInfo(10000, 100, 90, 0)
	sdk.SetCheckResult(true, 3, 10)
	testKp, err := account.NewKeyPair(nil, crypto.Ed25519)
	if err != nil {
		panic(err)
	}
	testID := "testID"
	err = sdk.CreateNewAccount(testID, testKp, 100000, 10000, 100000)
	if err != nil {
		panic(err)
	}
	sdk.SetAccount(testID, testKp)

	_, txHash, err := sdk.PublishContract(codePath, abiPath, "", false, "")
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Duration(50) * time.Second)
	client := rpc.NewApisClient(conns[0])
	resp, err := client.GetTxReceiptByTxHash(context.Background(), &rpc.HashReq{Hash: txHash})
	if err != nil {
		panic(err)
	}
	if tx.StatusCode(resp.TxReceipt.Status.Code) != tx.Success {
		panic("publish contract fail " + (resp.TxReceipt.String()))
	}
	return "Contract" + txHash
}

func main() {

	var iterNum = 800
	var parallelNum = 10
	initConn(parallelNum)

	contractID = publish()

	start := time.Now()

	for i := 0; i < iterNum; i++ {
		fmt.Println(i)
		transParallel(parallelNum)
	}

	fmt.Println("done. timecost=", time.Since(start))

}
