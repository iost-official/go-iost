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
var rootKey = "1rANSfcRzr4HkhbUFZ7L1Zp69JZZHiDDq5v7dNSbbEqeU4jxy3fszV4HGiaLQEyqVpS1dKT9g7zCVRxBVzuiUzB"
var contractID string

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

func sendTx(stx *tx.Tx, i int) ([]byte, error) {
	client := rpc.NewApisClient(conns[i])
	resp, err := client.SendRawTx(context.Background(), &rpc.RawTxReq{Data: stx.Encode()})
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
	//action := tx.NewAction("iost.system", "Transfer", `["IOSTfQFocqDn7VrKV7vvPqhAQGyeFU9XMYo5SNn5yQbdbzC75wM7C","IOSTgw6cmmWyiW25TMAK44N9coLCMaygx5eTfGVwjCcriEWEEjK2H",1]`)
	tmpAccount, _ := account.NewKeyPair(nil, crypto.Ed25519)
	action := tx.NewAction(contractID, "bet", fmt.Sprintf("[\"%s\",%d,%d,%d]", tmpAccount.ID, i%10, 100000000, 1))
	acc, _ := account.NewKeyPair(loadBytes(rootKey), crypto.Ed25519)
	trx := tx.NewTx([]*tx.Action{&action}, []string{}, 1000+int64(i), 1, time.Now().Add(time.Second*time.Duration(10000)).UnixNano(), 0)
	stx, err := tx.SignTx(trx, acc.ID, acc)
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
	codePath := "../../vm/test_data/lucky_bet.js"
	abiPath := codePath + ".abi"
	_, txHash, err := iwallet.PublishContract(codePath, abiPath, "", acc, 5, make([]string, 0), 10000, 1, false, "", true)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Duration(5) * time.Second)
	client := rpc.NewApisClient(conns[0])
	resp, err := client.GetTxReceiptByTxHash(context.Background(), &rpc.HashReq{Hash: common.Base58Encode(txHash)})
	if err != nil {
		panic(err)
	}
	if tx.StatusCode(resp.TxReceiptRaw.Status.Code) != tx.Success {
		panic("publish contract fail " + (resp.TxReceiptRaw.String()))
	}
	return "Contract" + common.Base58Encode(txHash)
}

func main() {

	var iterNum = 800
	var parallelNum = 30
	initConn(parallelNum)

	contractID = publish()

	start := time.Now()

	for i := 0; i < iterNum; i++ {
		fmt.Println(i)
		transParallel(parallelNum)
	}

	fmt.Println("done. timecost=", time.Since(start))

}
