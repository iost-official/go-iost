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

	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/vm"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check .sc file",
	Long:  `review and check .sc file, code and limits will be printed`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(`Error: source file not given`)
			return
		}
		path := args[0]
		fd, err := ReadFile(path)
		if err != nil {
			fmt.Println("Read file failed: ", err.Error())
			return
		}

		var mTx tx.Tx
		mTx.Decode(fd)
		PrintTx(mTx)

	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func PrintTx(mTx tx.Tx) {
	signer := make([]string, len(mTx.Signs))
	for _, s := range mTx.Signs {
		signer = append(signer, string(vm.PubkeyToIOSTAccount(s.Pubkey)))
	}
	var publisher string
	if mTx.Publisher.Pubkey == nil {
		publisher = ""
	} else {
		publisher = string(vm.PubkeyToIOSTAccount(mTx.Publisher.Pubkey))

	}
	fmt.Printf(`Transaction : 
Time: %v
Nonce: %v
contract:
    Price: %v
    Gas limit: %v
Code:
----
%v
----
Signer: %v
Publisher: %v 
`, mTx.Time, mTx.Nonce, mTx.Contract.Info().Price, mTx.Contract.Info().GasLimit, mTx.Contract.Code(), signer, publisher)
}
