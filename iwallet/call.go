// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iwallet

import (
	"fmt"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/spf13/cobra"
)

// TODO later
// var signers []string

//TODO refine the reminder here
// callCmd represents the compile command
var callCmd = &cobra.Command{
	Use:   "call",
	Short: "Call a method in some contract",
	Long: `Call a method in some contract
	the format of this command is:iwallet call contract_name0 function_name0 parameters0 contract_name1 function_name1 parameters1 ...
	(you can call more than one function in this command)
	the parameters is a string whose format is: ["arg0","arg1",...]
	example:./iwallet call "token.iost" "Transfer" '["iost","user0001","user0002","123.45",""]'
	`,
	Run: func(cmd *cobra.Command, args []string) {
		argc := len(args)
		if argc%3 != 0 {
			fmt.Println(`Error: number of args should be a multiplier of 3`)
			return
		}
		var actions []*tx.Action = make([]*tx.Action, argc/3)
		for i := 0; i < len(args); i += 3 {
			act := tx.NewAction(args[i], args[i+1], args[i+2]) //check sth here
			actions[i] = act
		}
		trx, err := sdk.createTx(actions)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		err = sdk.loadAccount()
		if err != nil {
			fmt.Printf("load account err %v\n", err)
			return
		}
		stx, err := sdk.signTx(trx)
		if err != nil {
			fmt.Printf("sign tx error %v\n", err)
			return
		}
		var txHash string
		fmt.Printf("sending tx %v", stx)
		txHash, err = sdk.sendTx(stx)
		if err != nil {
			fmt.Printf("send tx error %v\n", err)
			return
		}
		fmt.Println("send tx done")
		fmt.Println("the transaction hash is:", txHash)
		if sdk.checkResult {
			sdk.checkTransaction(txHash)
		}
	},
}

func init() {
	rootCmd.AddCommand(callCmd)
}
