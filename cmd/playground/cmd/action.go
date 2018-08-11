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

package cmd

import (
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"github.com/iost-official/Go-IOS-Protocol/new_vm"
	"github.com/spf13/cobra"
)

// actionCmd represents the action command
var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "Push action to playground",
	Long: `Push action to playground
Usage:
	playground action <contract> <action> '[arg1, arg2, ...]'
`,
	Run: func(cmd *cobra.Command, args []string) {
		bh, err := LoadBlockhead("block_info.json")
		if err != nil {
			panic(err)
		}
		db := NewDatabase()
		err = db.Load("state.json")
		if err != nil {
			panic(err)
		}
		eg := new_vm.NewEngine(bh, db)

		action := tx.NewAction(args[0], args[1], args[2])

		tx0, err := LoadTxInfo("tx_info.json")
		tx0.Actions = []tx.Action{action}

		fmt.Println(eg.Exec(tx0))

		db.Save("after.json")
	},
}

func init() {
	rootCmd.AddCommand(actionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// actionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// actionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
