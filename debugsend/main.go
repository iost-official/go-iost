package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/iost-official/prototype/account"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/iwallet/cmd"
	pb "github.com/iost-official/prototype/rpc"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"google.golang.org/grpc"
)

var server string = "127.0.0.1:30303"

func main() {
	rawCode := `
--- main 合约主入口
-- server1转账server2
-- @gas_limit 10000
-- @gas_price 0.001
-- @param_cnt 0
-- @return_cnt 0
function main()
	Transfer("2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT","mSS7EdV7WvBAiv7TChww7WE3fKDkEYRcVguznbQspj4K",0.5)
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
	for mtx.Nonce != 0 {
		mtx.Nonce++
		fmt.Println(mtx.Nonce)
		stx, err := tx.SignTx(mtx, acc)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		err = sendTx(stx)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func sendTx(stx tx.Tx) error {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewCliClient(conn)
	resp, err := client.PublishTx(context.Background(), &pb.Transaction{Tx: stx.Encode()})
	if err != nil {
		return err
	}
	switch resp.Code {
	case 0:
		return nil
	case -1:
		return errors.New("tx rejected")
	default:
		return errors.New("unknown return")
	}
}
