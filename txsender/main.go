package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/tx"
	pb "github.com/iost-official/prototype/rpc"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var log = logrus.New()

var acc string = "2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT"

var (
	accId   = flag.Int("account", 0, "account_id")
	money   = flag.Int("money", 1, "money")
	tps     = flag.Int("tps", 10, "tps you want")
	cluster = flag.String("cluster", "testnet", "cluster name, example: test, testnet, local")
)

var servers = map[string][]string{
	"testnet": []string{
		"18.179.143.193:30303",
		"52.56.118.10:30303",
		"13.228.206.188:30303",

		"13.232.96.221:30303",
		"18.184.239.232:30303",
		"13.124.172.86:30303",
		"52.60.163.60:30303",
	},
	"test": []string{
		"13.236.207.159:30303",
		"13.236.209.209:30303",
		"54.206.55.116:30303",
		"54.206.49.230:30303",
		"13.236.177.85:30303",
		"13.236.153.25:30303",
		"13.211.188.83:30303",
	},
	"local": []string{
		"127.0.0.1:30303",
		"127.0.0.1:30313",
		"127.0.0.1:30323",
	},
}

var server_num = map[string]int{
	"testnet": 7,
	"test":    7,
	"local":   3,
}

var accounts []string = []string{
	"2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT",
	"tUFikMypfNGxuJcNbfreh8LM893kAQVNTktVQRsFYuEU",
	"s1oUQNTcRKL7uqJ1aRqUMzkAkgqJdsBB7uW9xrTd85qB",
	"22zr9ows3qndmAjnkiPFex26taATEaEfjGkatVCr5akSU",
	"wSKjLjqWbhH2LcJFwTW9Nfq9XPdhb4pw9KCM7QGtemZG",
	"oh7VBi17aQvG647cTfhhoRGby3tH55o3Qv7YHWD5q8XU",
	"28mKnLHaVvc1YRKc9CWpZxCpo2gLVCY3RL5nC9WbARRym",
}

func LoadBytes(s string) []byte {
	buf := common.Base58Decode(s)
	return buf
}

func send(wg *sync.WaitGroup, mtx tx.Tx, acc account.Account, startNonce int64, routineId int) {
	defer wg.Done()
	log.Info("cluster: %v, routineId: %v, server_num: %v", *cluster, routineId, server_num[*cluster])
	conn, err := grpc.Dial(servers[*cluster][(routineId%server_num[*cluster])], grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	pclient := pb.NewCliClient(conn)

	for i := startNonce; i != -1; i++ {

		mtx.Nonce = i
		log.Debugf("Now Nonce: %v", mtx.Nonce)
		mtx.Time = time.Now().UnixNano()
		stx, err := tx.SignTx(mtx, acc)
		if err != nil {
			log.Errorf("Sign transaction error:", err)
			return
		}

		err = sendTx(&stx, pclient)
		if err != nil {
			log.Errorf("Send transaction error:", err)
			return
		}

	}
	return
}

func main() {
	flag.Parse()
	if accId == nil || money == nil || tps == nil {
		return
	}
	acc = accounts[*accId]
	rawCode := `
--- main 合约主入口
-- server1转账server2
-- @gas_limit 10000
-- @gas_price 0.001
-- @param_cnt 0
-- @return_cnt 0
function main()
	Transfer("` + acc + `","mSS7EdV7WvBAiv7TChww7WE3fKDkEYRcVguznbQspj4K",` + strconv.Itoa(*money) + `)
end--f
`
	var contract vm.Contract
	parser, _ := lua.NewDocCommentParser(rawCode)
	contract, err := parser.Parse()
	if err != nil {
		log.Fatalf("Contract parse error:", err)
	}
	mtx := tx.NewTx(1, contract)
	acc, err := account.NewAccount(LoadBytes("BRpwCKmVJiTTrPFi6igcSgvuzSiySd7Exxj7LGfqieW9"))
	if err != nil {
		log.Fatalf("New account error:", err)
	}
	//test time of sending one tx
	var curtps float64 = 0.0
	for i := 0; i < 3; i++ {
		start := time.Now().UnixNano()
		stx, err := tx.SignTx(mtx, acc)
		if err != nil {
			log.Errorf("Sign transaction error:", err)
			return
		}
		conn, err := grpc.Dial(servers[*cluster][0], grpc.WithInsecure())
		if err != nil {
			return
		}
		pclient := pb.NewCliClient(conn)
		sendTx(&stx, pclient)
		conn.Close()
		end := time.Now().UnixNano()
		curtps += float64(end - start)
	}
	curtps /= 3
	curtps = float64(1e9) / curtps
	routineNum := int(float64(*tps) / curtps)
	if routineNum < 1 {
		routineNum = 1
	} else if routineNum > 5000 {
		routineNum = 5000
	}
	fmt.Printf("number of routines: %d", routineNum)
	var wg sync.WaitGroup
	wg.Add(routineNum)
	for i := 0; i < routineNum; i++ {
		go send(&wg, mtx, acc, int64(i)*int64(10000000000), i)
	}
	wg.Wait()
	log.Fatal("Func main finished")
}

func sendTx(stx *tx.Tx, pclient pb.CliClient) error {
	//resp, err := client.PublishTx(context.Background(), &pb.Transaction{Tx: stx.Encode()})
	_, err := pclient.PublishTx(context.Background(), &pb.Transaction{Tx: stx.Encode()})
	if err != nil {
		return err
	}
	return nil
	/*
		switch resp.Code {
		case 0:
			return nil
		case -1:
			return errors.New("tx rejected")
		default:
			return errors.New("unknown return")
		}
	*/
}
