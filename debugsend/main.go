package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/iwallet/cmd"
	pb "github.com/iost-official/prototype/rpc"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"google.golang.org/grpc"
)

var acc string = "2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT"
var servers []string = []string{
	"127.0.0.1",
	"18.179.143.193",
	"52.56.118.10",
	"13.228.206.188",
	"13.232.96.221",
	"18.184.239.232",
	"13.124.172.86",
	"52.60.163.60",
}
var servers1 []string = []string{
        "127.0.0.1",
        "13.236.207.159",
        "13.236.209.209",
        "54.206.55.116",
        "54.206.49.230",
        "13.236.177.85",
        "13.236.153.25",
        "13.211.188.83",
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

func send(wg *sync.WaitGroup, mtx tx.Tx, acc account.Account, startNonce int64, routineId int) {
	defer wg.Done()
	conn, err := grpc.Dial(servers[(routineId%7)+1]+":30303", grpc.WithInsecure())
	if err != nil {
		return 
	}
	defer conn.Close()
	pclient := pb.NewCliClient(conn)
	
	for i := startNonce; i != -1; i++ {
		mtx.Nonce = i
		fmt.Println(mtx.Nonce)
		mtx.Time = time.Now().UnixNano()
		stx, err := tx.SignTx(mtx, acc)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		err = sendTx(&stx,pclient)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	return
}
func main() {
	accId := flag.Int("account", 0, "account_id")
	money := flag.Int("money", 1, "money")
	nums := flag.Int("routines", 10, "number of routines")
	flag.Parse()
	if accId == nil || money == nil || nums == nil {
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
		fmt.Println(err.Error())
		return
	}
	mtx := tx.NewTx(1, contract)
	acc, err := account.NewAccount(cmd.LoadBytes("BRpwCKmVJiTTrPFi6igcSgvuzSiySd7Exxj7LGfqieW9"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < *nums; i++ {
		go send(&wg, mtx, acc, int64(i)*int64(10000000000), i)
	}
	wg.Wait()
	fmt.Println("main")
}

func sendTx(stx *tx.Tx,pclient pb.CliClient) error {
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
