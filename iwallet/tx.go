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
	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/spf13/cobra"
)

// txCmd save a contract with given actions to a file
var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "save a contract with given actions to a file",
	Long:  `save a contract with given actions to a file, command line is similar to 'call'' command`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if outputFile == "" {
			return fmt.Errorf("--output should not be empty")
		}
		var actions []*rpcpb.Action
		actions, err = actionsFromFlags(args)
		if err != nil {
			return
		}
		trx, err := sdk.createTx(actions)
		if err != nil {
			return err
		}
		return saveProto(trx, outputFile)
	},
}

func init() {
	rootCmd.AddCommand(txCmd)
	txCmd.Flags().StringVarP(&outputFile, "output", "", "", "output file name to write signature")
}
