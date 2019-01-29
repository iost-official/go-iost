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

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/spf13/cobra"
)

// systemCmd represents the system command.
var systemCmd = &cobra.Command{
	Use:     "system",
	Aliases: []string{"sys"},
	Short:   "Send system contract action to blockchain",
	Long:    `Send system contract action to blockchain`,
}

func actionSender(contract, method string, methodArgs ...interface{}) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		methodArgsBytes, err := json.Marshal(methodArgs)
		if err != nil {
			return err
		}
		action := NewAction(contract, method, string(methodArgsBytes))
		actions := []*rpcpb.Action{action}
		tx, err := sdk.createTx(actions)
		if err != nil {
			return fmt.Errorf("failed to create tx: %v", err)
		}
		err = sdk.LoadAccount()
		if err != nil {
			return fmt.Errorf("failed to load account: %v", err)
		}
		_, err = sdk.SendTx(tx)
		return err
	}
}

func init() {
	rootCmd.AddCommand(systemCmd)
}
