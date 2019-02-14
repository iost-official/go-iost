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
	"github.com/iost-official/go-iost/sdk"

	"github.com/spf13/cobra"
)

// transactionCmd represents the transaction command.
var transactionCmd = &cobra.Command{
	Use:     "transaction transactionHash",
	Aliases: []string{"tx"},
	Short:   "Find transactions",
	Long:    `Find transaction by transaction hash`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmd.Usage()
			return fmt.Errorf("please enter the transaction hash")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		txRaw, err := iwalletSDK.GetTxByHash(args[0])
		if err != nil {
			return err
		}
		fmt.Println(sdk.MarshalTextString(txRaw))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(transactionCmd)
}
