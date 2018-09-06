package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/crypto"
	pb "github.com/iost-official/Go-IOS-Protocol/rpc"
	"google.golang.org/grpc"
)

var conns []*grpc.ClientConn

func initConn(num int) {
	conns = make([]*grpc.ClientConn, num)
	allServers := []string{"13.237.151.211:30002", "35.177.202.166:30002", "18.136.110.166:30002",
		"13.232.76.188:30002", "52.59.86.255:30002", "54.180.13.100:30002", "35.183.163.183:30002"}
	// allServers := []string{"192.168.1.13:30302", "192.168.1.13:30305", "192.168.1.13:30308"}
	for i := 0; i < num; i++ {
		conn, err := grpc.Dial(allServers[i%7], grpc.WithInsecure())
		if err != nil {
			continue
		}
		conns[i] = conn
	}

}

func transParallel(num int) {
	if conns == nil {
		initConn(num)
	}
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
	if conns[i] == nil {
		return nil, errors.New("nil conn")
	}
	client := pb.NewApisClient(conns[i])
	resp, err := client.SendRawTx(context.Background(), &pb.RawTxReq{Data: stx.Encode()})
	if err != nil {
		return nil, err
	}
	return []byte(resp.Hash), nil
	/*
		switch resp.Code {
		case 0:
			return resp.Hash, nil
		case -1:
			return nil, errors.New("tx rejected")
		default:
			return nil, errors.New("unknown return")
		}
	*/
}

func loadBytes(s string) []byte {
	if s[len(s)-1] == 10 {
		s = s[:len(s)-1]
	}
	buf := common.Base58Decode(s)
	return buf
}

func transfer(i int) {
	action := tx.NewAction("iost.system", "Transfer", `["IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP","IOST24jsSGj2WxSRtgZkCDng19LPbT48HMsv2Nz13NXEYoqR1aYyvS",1]`)
	acc, _ := account.NewAccount(loadBytes("5ifJUpGWJ69S2eKsKYLDcajVxrc5yZk2CD7tJ29yK6FyjAtmeboK3G4Ag5p22uZTijBP3ftEDV4ymXZF1jGqu9j4"), crypto.Ed25519)
	// fmt.Println(acc.Pubkey, account.GetIDByPubkey(acc.Pubkey))
	trx := tx.NewTx([]*tx.Action{&action}, [][]byte{}, 1000, 1, time.Now().Add(time.Second*time.Duration(10000)).UnixNano())
	stx, err := tx.SignTx(trx, acc)
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

func main() {

	var num = 500000

	start := time.Now()

	for i := 0; i < num; i++ {
		transParallel(49)
	}

	fmt.Println("done. timecost=", time.Since(start))

}
