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

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
)

// signCmd represents the command used to sign a transaction.
var signCmd = &cobra.Command{
	Use:   "sign txFile keyFile outputFile",
	Short: "Sign a tx and save the signature",
	Long:  `Sign a transaction loaded from given txFile with keyFile(account json file or private key file) and save the signature as outputFile`,
	Example: `  iwallet sign tx.json ~/.iwallet/test0.json sign.json
  iwallet sign tx.json ~/.iwallet/test0_ed25519 sign.json`,
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
			return fmt.Errorf("failed to load transaction file %v: %v", txFile, err)
		}
		accInfo, err := loadAccountFromFile(signKeyFile, true)
		if err != nil {
			return fmt.Errorf("failed to load account from file %v: %v", signKeyFile, err)
		}
		kp, err := accInfo.Keypairs["active"].toKeyPair()
		if err != nil {
			return fmt.Errorf("failed to get key pair from file %v: %v", signKeyFile, err)
		}
		sig := sdk.GetSignatureOfTx(trx, kp)
		if verbose {
			fmt.Println("Signature:")
			fmt.Println(sdk.MarshalTextString(sig))
		}
		err = sdk.SaveProtoStructToJSONFile(sig, outputFile)
		if err != nil {
			return fmt.Errorf("failed to save signature as file %v: %v", outputFile, err)
		}
		fmt.Println("Successfully saved signature as:", outputFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
}
