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

var outputFile string
var signKeyFile string

// signCmd sign a tx, and save the signature to a binary file
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "sign a tx, and save the signature to a binary file",
	Long:  "create a tx from command similar to `call` command, then sign it and save the signature to file",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if outputFile == "" {
			return fmt.Errorf("output file name should be provided with --output flag")
		}
		var actions []*rpcpb.Action
		actions, err = actionsFromFlags(args)
		if err != nil {
			return err
		}
		trx, err := sdk.createTx(actions)
		if err != nil {
			return err
		}
		kp, err := sdk.loadKeyPair(signKeyFile)
		if err != nil {
			return err
		}
		sig := sdk.getSignatureOfTx(trx, kp)
		err = saveTo(outputFile, signatureToBytes(sig))
		if err != nil {
			return err
		}
		fmt.Printf("save signature to %v done", outputFile)
		return
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVarP(&outputFile, "output", "", "", "output file name to write signature")
	signCmd.Flags().StringVarP(&signKeyFile, "sign_key", "", "", "key file used for signing")
}
