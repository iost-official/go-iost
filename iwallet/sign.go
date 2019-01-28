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
var txFile string

// signCmd represents the command used to sign a transaction.
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a tx loaded from given file and save the signature as a binary file",
	Long:  `Sign a tx loaded from given file (--tx_file) and save the signature as a binary file (--output)`,
	Args: func(cmd *cobra.Command, args []string) error {
		if outputFile == "" {
			cmd.Usage()
			return fmt.Errorf("output file name should be provided with --output flag")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		trx := &rpcpb.TransactionRequest{}
		err := loadProto(txFile, trx)
		if err != nil {
			return fmt.Errorf("failed to load transaction file: %v", err)
		}
		kp, err := loadKeyPair(signKeyFile, sdk.GetSignAlgo())
		if err != nil {
			return fmt.Errorf("failed to load key pair: %v", err)
		}
		sig := sdk.getSignatureOfTx(trx, kp)
		err = saveProto(sig, outputFile)
		if err != nil {
			return fmt.Errorf("failed to save signature: %v", err)
		}
		fmt.Println("Successfully saved signature as:", outputFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVarP(&outputFile, "output", "", "", "output file name to write signature")
	signCmd.Flags().StringVarP(&signKeyFile, "sign_key", "", "", "key file used for signing")
	signCmd.Flags().StringVarP(&txFile, "tx_file", "", "", "load tx from this file")
}
