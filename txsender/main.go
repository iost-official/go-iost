package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/tx"
	pb "github.com/iost-official/prototype/rpc"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var log = logrus.New()

var acc string = "2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT"

var (
	accId   = flag.Int("account", 0, "money sender in the contract")
	money   = flag.Int("money", 1, "money you send in one tx")
	tps     = flag.Int("tps", 10, "txs per second you send")
	cluster = flag.String("cluster", "testnet", "cluster name, example: test, testnet, local")
	rout    = flag.Int("routines", -1, "number of routines you create")
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
		"54.79.77.101:30303",
		"13.236.209.209:30303",
		"54.206.55.116:30303",
		"54.206.49.230:30303",
		"13.236.177.85:30303",
		"13.236.153.25:30303",
		"13.211.188.83:30303",
	},
	"california": []string{
		"13.56.255.143:30303",
		"18.144.11.65:30303",
		"13.57.176.233:30303",
		"54.183.115.79:30303",
		"13.56.223.196:30303",
		"18.144.42.61:30303",
		"13.57.185.234:30303",
	},

	"local": []string{
		"127.0.0.1:30303",
		"127.0.0.1:30313",
		"127.0.0.1:30323",
	},
}

var server_num = map[string]int{
	"testnet":    7,
	"test":       7,
	"california": 7,
	"local":      3,
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

var needsleep bool = false
var duration int = 0
var eps float64 = (1e-6)

func send(wg *sync.WaitGroup, mtx tx.Tx, acc account.Account, startNonce int64, routineId int) {
	defer wg.Done()
	log.Infof("cluster: %v, routineId: %v, server_num: %v", *cluster, routineId, server_num[*cluster])

	for i := startNonce; i != -1; i++ {
		start := time.Now().UnixNano()
		mtx.Nonce = i
		log.Debugf("Now Nonce: %v", mtx.Nonce)
		mtx.Time = time.Now().UnixNano()
		stx, err := tx.SignTx(mtx, acc)
		if err != nil {
			log.Errorf("Sign transaction error:", err)
			return
		}

		err = sendTx(&stx)
		if err != nil {
			log.Errorf("Send transaction error:", err)
			continue
		}
		end := time.Now().UnixNano()
		curtps := float64(1e9) / float64(end-start)
		if needsleep {
			if curtps > float64(*tps)+eps {
				duration += 1000
			} else if curtps < float64(*tps)-eps {
				duration -= 1000
			}
			if duration < 0 {
				duration = 0
			}
			time.Sleep(time.Duration(duration))
		}
	}
	return
}

func main() {
	flag.Parse()
	if accId == nil || money == nil || tps == nil {
		return
	}
	rand.Seed(time.Now().UnixNano())
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
	var routineNum int
	if *rout == -1 {
		//test time of sending one tx
		var curtps float64 = 0.0
		for i := 0; i < 3; i++ {
			start := time.Now().UnixNano()
			stx, err := tx.SignTx(mtx, acc)
			if err != nil {
				log.Errorf("Sign transaction error:", err)
				return
			}
			sendTx(&stx)
			end := time.Now().UnixNano()
			curtps += float64(end - start)
		}
		curtps /= 3
		curtps = float64(1e9) / curtps
		routineNum = int(float64(*tps) / curtps)
		if routineNum < 1 {
			routineNum = 1
		} else if routineNum > 5000 {
			routineNum = 5000
		}
		if routineNum == 1 && curtps > float64(*tps) {
			needsleep = true
		}
	} else {
		routineNum = *rout
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

func sendTx(stx *tx.Tx) error {
	id := rand.Intn(server_num[*cluster] - 1)
	fmt.Printf("sendto: %v\n", id)
	conn, err := grpc.Dial(servers[*cluster][id], grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("conn err")
	}
	defer conn.Close()
	pclient := pb.NewCliClient(conn)

	//resp, err := client.PublishTx(context.Background(), &pb.Transaction{Tx: stx.Encode()})
	_, err = pclient.PublishTx(context.Background(), &pb.Transaction{Tx: stx.Encode()})
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
