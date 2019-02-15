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
	"time"

	"github.com/spf13/cobra"
)

// saveCmd would save a transaction request with given actions to a file.
var saveCmd = &cobra.Command{
	Use:   "save [ACTION]...",
	Short: "Save a transaction request with given actions to a file",
	Long: `Save a transaction request with given actions to a file
	Would accept arguments as call actions and create a transaction request from them.
	An ACTION is a group of 3 arguments: contract name, function name, method parameters.
	The method parameters should be a string with format '["arg0","arg1",...]'.`,
	Example: `  iwallet save "token.iost" "transfer" '["iost","user0001","user0002","123.45",""]' -o tx.json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if outputFile == "" {
			cmd.Usage()
			return fmt.Errorf("output file name should be provided with --output flag")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		actions, err := actionsFromFlags(args)
		if err != nil {
			return err
		}
		trx, err := iwalletSDK.CreateTxFromActions(actions)
		if err != nil {
			return err
		}
		t, err := parseTimeFromStr(txTime)
		if err != nil {
			return err
		}
		trx.Time = t
		trx.Expiration = trx.Time + expiration*1e9
		return sdk.SaveProtoStructToJSONFile(trx, outputFile)
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
	saveCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file to save transaction request")
}

func parseTimeFromStr(s string) (int64, error) {
	var t time.Time
	if s == "" {
		return time.Now().UnixNano(), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, fmt.Errorf("invalid time %v, should in format %v", s, time.RFC3339)
	}
	return t.UnixNano(), nil
}

var txTime string
