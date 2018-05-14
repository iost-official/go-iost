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

	"io/ioutil"
	"os"

	"github.com/iost-official/prototype/core/tx"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(`Error: source file not given`)
			return
		}
		path := args[0]
		fi, err := os.Open(path)
		if err != nil {
			fmt.Println("Error: input file not found")
		}
		defer fi.Close()
		fd, err := ioutil.ReadAll(fi)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		var mTx tx.Tx
		mTx.Decode(fd)
		fmt.Printf(`Transaction : 
Time: %v
Nonce: %v
Contract:
    Price: %v
    Gas limit: %v
Code:
----
%v
----
`, mTx.Time, mTx.Nonce, mTx.Contract.Info().Price, mTx.Contract.Info().GasLimit, mTx.Contract.Code())

	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
