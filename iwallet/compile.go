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
	"os"

	"time"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile contract files to smart contract",
	Long:  `Compile contract files to smart contract. `,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println(`Error: source code file or abi file not given`)
			return
		}
		codePath := args[0]
		fd, err := readFile(codePath)
		if err != nil {
			fmt.Println("Read source code file failed: ", err.Error())
			return
		}
		code := string(fd)
		abiPath := args[1]
		fd, err = readFile(abiPath)
		if err != nil {
			fmt.Println("Read abi file failed: ", err.Error())
			return
		}
		abi := string(fd)

		compiler := new(contract.Compiler)
		if compiler == nil {
			fmt.Println("gen compiler instance failed")
			return
		}
		contract, err := compiler.Parse("", code, abi)
		if err != nil {
			fmt.Printf("gen contract error:%v\n", err)
			return
		}
		action := tx.NewAction("iost.system", "SetCode", `["`+contract.B64Encode()+`",]`)
		pubkeys := make([][]byte, len(signers))
		for i, pubkey := range signers {
			pubkeys[i] = loadBytes(string(pubkey))
		}

		trx := tx.NewTx([]*tx.Action{&action}, pubkeys, gasLimit, gasPrice, time.Now().Add(time.Second*time.Duration(expiration)).UnixNano())

		if len(signers) == 0 {
			fmt.Println("you don't indicate any signers,so this tx will be sent to the iostNode directly")
			fsk, err := readFile(kpPath)
			if err != nil {
				fmt.Println("Read file failed: ", err.Error())
				return
			}

			acc, err := account.NewAccount(loadBytes(string(fsk)), getSignAlgo(signAlgo))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			stx, err := tx.SignTx(trx, acc)
			var txHash []byte
			txHash, err = sendTx(stx)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("ok")
			fmt.Println(saveBytes(txHash))
			return
		}

		bytes := trx.Encode()

		if dest == "default" {
			dest = changeSuffix(args[0], ".sc")
		}

		err = saveTo(dest, bytes)
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}

var dest string
var gasLimit int64
var gasPrice int64
var expiration int64
var signers []string

func init() {
	rootCmd.AddCommand(compileCmd)
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	compileCmd.Flags().Int64VarP(&gasLimit, "gaslimit", "l", 1000, "gasLimit for a transaction")
	compileCmd.Flags().Int64VarP(&gasPrice, "gasprice", "p", 1, "gasPrice for a transaction")
	compileCmd.Flags().Int64VarP(&expiration, "expiration", "e", 0, "expiration timestamp for a transaction")
	compileCmd.Flags().StringSliceVarP(&signers, "signers", "", []string{}, "signers who should sign this transaction")
	compileCmd.Flags().StringVarP(&kpPath, "key-path", "k", home+"/.iwallet/id_ed25519", "Set path of sec-key")
	compileCmd.Flags().StringVarP(&signAlgo, "signAlgo", "a", "ed25519", "Sign algorithm")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compi leCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
