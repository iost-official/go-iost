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

	"os"
	"strings"

	"github.com/iost-official/prototype/account"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Account manage",
	Long:  `Manage account of local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case nickName != "no":
			if strings.ContainsAny(nickName, `?*:|/\"`) || len(nickName) > 16 {
				fmt.Println("invalid nick name")
			}
			ac, _ := account.NewAccount(nil)
			pubfile, err := os.Create(kvPath + "/" + nickName + "_secp.pub")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer pubfile.Close()

			_, err = pubfile.WriteString(SaveBytes(ac.Pubkey))
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			secFile, err := os.Create(kvPath + "/" + nickName + "_secp")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer secFile.Close()

			_, err = secFile.WriteString(SaveBytes(ac.Seckey))
			if err != nil {
				fmt.Println(err.Error())
				return
			}

		default:
			fmt.Println("invalid input")
		}
	},
}

var kvPath string
var nickName string

func init() {
	rootCmd.AddCommand(accountCmd)

	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	accountCmd.Flags().StringVarP(&nickName, "create", "c", "id", "Create new account, using input as nickname")
	accountCmd.Flags().StringVarP(&kvPath, "path", "p", home+"/.ssh", "Set path of key pair file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
