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

// publishCmd represents the publish command.
var publishCmd = &cobra.Command{
	Use:     "publish codePath abiPath [contractID [updateID]]",
	Aliases: []string{"pub"},
	Short:   "Publish a contract",
	Long:    `Publish a contract by a contract and an abi file`,
	Example: `  iwallet publish ./example.js ./example.js.abi --account test0
  iwallet publish ./example.js ./example.js.abi -u ContractXXX --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			cmd.Usage()
			return fmt.Errorf("please enter the code path and the abi path")
		}
		if update && len(args) < 3 {
			cmd.Usage()
			return fmt.Errorf("please enter the contract id")
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		codePath := args[0]
		abiPath := args[1]

		conID := ""
		if update {
			conID = args[2]
		}
		updateID := ""
		if update && len(args) >= 4 {
			updateID = args[3]
		}

		err := sdk.LoadAccount()
		if err != nil {
			return fmt.Errorf("failed to load account: %v", err)
		}
		_, txHash, err := sdk.PublishContract(codePath, abiPath, conID, update, updateID)
		if err != nil {
			return fmt.Errorf("failed to create tx: %v", err)
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
