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
	"github.com/spf13/cobra"

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
)

// sendCmd represents the send command that send a contract with given actions.
var sendCmd = &cobra.Command{
	Use:     "send txFile",
	Short:   "Send transaction onto blockchain by given json file",
	Long:    `Send transaction onto blockchain by given json file`,
	Example: `iwallet send tx.json --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		tx := &rpcpb.TransactionRequest{}
		err := sdk.LoadProtoStructFromJSONFile(args[0], tx)
		if err != nil {
			return err
		}
		return sendTx(tx)
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
