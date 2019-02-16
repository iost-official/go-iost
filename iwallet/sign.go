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

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/spf13/cobra"
)

var outputFile string
var signKeyFile string
var txFile string

// signCmd represents the command used to sign a transaction.
var signCmd = &cobra.Command{
	Use:     "sign txFile keyFile outputFile",
	Short:   "Sign a tx and save the signature",
	Long:    `Sign a tx loaded from given file with private key file and save the signature`,
	Example: `  iwallet sign tx.json ~/.iwallet/test0_ed25519 sign.json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "txFile", "keyFile", "outputFile"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		txFile := args[0]
		signKeyFile := args[1]
		outputFile := args[2]

		trx := &rpcpb.TransactionRequest{}
		err := sdk.LoadProtoStructFromJSONFile(txFile, trx)
		if err != nil {
			return fmt.Errorf("failed to load transaction file: %v", err)
		}
		kp, err := sdk.LoadKeyPair(signKeyFile, signAlgo)
		if err != nil {
			return fmt.Errorf("failed to load key pair: %v", err)
		}
		sig := sdk.GetSignatureOfTx(trx, kp)
		err = sdk.SaveProtoStructToJSONFile(sig, outputFile)
		if err != nil {
			return fmt.Errorf("failed to save signature: %v", err)
		}
		fmt.Println("Successfully saved signature as:", outputFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
}
