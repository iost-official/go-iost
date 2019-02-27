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

	"github.com/iost-official/go-iost/sdk"
)

// saveCmd would save a transaction request with given actions to a file.
var saveCmd = &cobra.Command{
	Use:   "save [ACTION]...",
	Short: "Save a transaction request with given actions to a file",
	Long: `Save a transaction request with given actions to a file
	Would accept arguments as call actions and create a transaction request from them.
	An ACTION is a group of 3 arguments: contract name, function name, method parameters.
	The method parameters should be a string with format '["arg0","arg1",...]'.`,
	Example: `  iwallet save "token.iost" "transfer" '["iost","user0001","user0002","123.45",""]' -j tx.json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if outputTxFile == "" {
			return errorWithHelp(cmd, "please provide output json file with flag --json/-j")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		actions, err := actionsFromFlags(args)
		if err != nil {
			return err
		}
		tx, err := initTxFromActions(actions)
		if err != nil {
			return err
		}
		err = sdk.SaveProtoStructToJSONFile(tx, outputTxFile)
		if err == nil {
			fmt.Println("Successfully saved transaction request as json file:", outputTxFile)
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
