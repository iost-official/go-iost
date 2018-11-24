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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
)

// receiptCmd represents the receipt command
var receiptCmd = &cobra.Command{
	Use:   "receipt",
	Short: "find receipt",
	Long:  `find receipt by transaction hash`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(`Error: transaction hash not given`)
			return
		}
		txReceipt, err := sdk.getTxReceiptByTxHash(args[0])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		ret, err := json.MarshalIndent(txReceipt, "", "    ")
		if err != nil {
			fmt.Printf("error %v\n", err)
			return
		}
		fmt.Println(string(ret))
	},
}

func init() {
	rootCmd.AddCommand(receiptCmd)
}
