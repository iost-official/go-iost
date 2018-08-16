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

package cmd

import (
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/spf13/cobra"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:    "compile",
	Short: "Compile contract files to smart contract",
	Long:  `Compile contract files to smart contract. `,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println(`Error: source code file or abi file not given`)
			return
		}
		codePath := args[0]
		fd, err := ReadFile(codePath)
		if err != nil {
			fmt.Println("Read source code file failed: ", err.Error())
			return
		}
		code := string(fd)
		abiPath := args[1]
		fd, err = ReadFile(abiPath)
		if err != nil {
			fmt.Println("Read abi file failed: ", err.Error())
			return
		}
		abi := string(fd)

		compiler:=new(contract.Compiler)
		if compiler==nil{
			fmt.Println("gen compiler instance failed")
			return
		}
		contract,err:=compiler.Parse("",code,abi)
		if err!=nil{
			fmt.Printf("gen contract error:%v\n",err)
			return
		}
		action:=tx.NewAction("iost.system","setcode",`["`+contract.Encode()+`",]`)
		pubkeys:=make([][]byte,len(signers))
		for i,pubkey:=range signers{
			pubkeys[i]=LoadBytes(string(pubkey))
		}
		trx:=tx.NewTx([]tx.Action{action,}, pubkeys, gasLimit, gasPrice, expiration)




		bytes := trx.Encode()

		if dest == "default" {
			dest = ChangeSuffix(args[0], ".sc")
		}

		err = SaveTo(dest, bytes)
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}

var dest string
var gasLimit uint64
var gasPrice uint64
var expiration int64
var signers []string
func init() {
	rootCmd.AddCommand(compileCmd)
	
	compileCmd.Flags().Uint64VarP(&gasLimit, "gaslimit", "l", 1000, "gasLimit for a transaction")
	compileCmd.Flags().Uint64VarP(&gasPrice, "gasprice", "p", 1, "gasPrice for a transaction")
	compileCmd.Flags().Int64VarP(&expiration, "expiration", "e", 0, "expiration timestamp for a transaction")
	compileCmd.Flags().StringSliceVarP(&signers, "signers", "s",[]string{} , "signers who should sign this transaction")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compi leCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
