package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	pb "github.com/iost-official/Go-IOS-Protocol/rpc"
	"google.golang.org/grpc"
)

func transParallel(num int) {
	wg := new(sync.WaitGroup)
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			transfer()
			wg.Done()

		}()
	}
	wg.Wait()
}

var GenesisAccount = map[string]int64{
	"IOST5FhLBhVXMnwWRwhvz5j9NyWpBSchAMzpSMZT21xZqT8w7icwJ5": 13400000000, // seckey:BCV7fV37aSWNx1N1Yjk3TdQXeHMmLhyqsqGms1PkqwPT
	"IOST6Jymdka3EFLAv8954MJ1nBHytNMwBkZfcXevE2PixZHsSrRkbR": 13200000000, // seckey:2Hoo4NAoFsx9oat6qWawHtzqFYcA3VS7BLxPowvKHFPM
	"IOST7gKuvHVXtRYupUixCcuhW95izkHymaSsgKTXGDjsyy5oTMvAAm": 13100000000, // seckey:6nMnoZqgR7Nvs6vBHiFscEtHpSYyvwupeDAyfke12J1N
}

func sendTx(stx tx.Tx) ([]byte, error) {
	allServers := []string{"127.0.0.1:30302", "127.0.0.1:30305", "127.0.0.1:30308"}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(3)
	conn, err := grpc.Dial(allServers[n], grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := pb.NewApisClient(conn)
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

func transfer() {
	action := tx.NewAction("iost.system", "Transfer", `["IOST5FhLBhVXMnwWRwhvz5j9NyWpBSchAMzpSMZT21xZqT8w7icwJ5","IOSTponZK9JJqZAsEWMF1BCZkSKnRP7abGbKjZb49nidfYW8",1]`)
	acc, _ := account.NewAccount(loadBytes("BCV7fV37aSWNx1N1Yjk3TdQXeHMmLhyqsqGms1PkqwPT"))
	// fmt.Println(acc.Pubkey, account.GetIDByPubkey(acc.Pubkey))
	trx := tx.NewTx([]*tx.Action{&action}, [][]byte{}, 1000, 1, time.Now().Add(time.Second*time.Duration(10000)).UnixNano())
	stx, err := tx.SignTx(trx, acc)
	//fmt.Println("verify", stx.VerifySelf())
	if err != nil {
		return
	}
	_, err = sendTx(*stx)
	if err != nil {
		return
	}
}

func main() {
	for {
		var num = 40
		start := time.Now()

		for i := 0; i < num; i++ {
			transParallel(10)
		}
		fmt.Println("done. timecost=", time.Since(start))
	}
}
