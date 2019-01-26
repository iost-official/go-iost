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

	"github.com/spf13/cobra"
)

var update bool

// publishCmd represents the compile command.
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a contract",
	Long: `Publish a contract by a contract and an abi file
	Example:
		iwallet publish ./example.js ./example.js.abi
		iwallet publish -u ./example.js ./example.js.abi contractID [updateID]
	`,

	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) < 2 {
			fmt.Println(`Usage: iwallet publish ./example.js ./example.js.abi`)
			return
		}
		codePath := args[0]
		abiPath := args[1]

		conID := ""
		if update {
			if len(args) < 3 {
				fmt.Println("Please enter the contract id")
				return
			}
			conID = args[2]
		}
		updateID := ""
		if update && len(args) >= 4 {
			updateID = args[3]
		}

		err = sdk.LoadAccount()
		if err != nil {
			fmt.Printf("Failed to load account: %v\n", err)
			return
		}
		_, txHash, err := sdk.PublishContract(codePath, abiPath, conID, update, updateID)
		if err != nil {
			fmt.Printf("Failed to create tx: %v\n", err)
			return
		}
		if sdk.checkResult {
			if err := sdk.checkTransaction(txHash); err != nil {
				return err
			}
			if !update {
				fmt.Println("The contract id is Contract" + txHash)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().BoolVarP(&update, "update", "u", false, "update contract")
}
