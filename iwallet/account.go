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
	"strings"

	"github.com/iost-official/go-iost/account"
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "KeyPair manage",
	Long:  `Manage account of local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Printf("new account name cannot be empty\n")
			return
		}
		newName := args[0]
		if strings.ContainsAny(newName, `?*:|/\"`) || len(newName) > 16 {
			fmt.Printf("invalid account name\n")
			return
		}
		algo := sdk.getSignAlgo()
		newKp, err := account.NewKeyPair(nil, algo)
		if err != nil {
			fmt.Printf("create key pair failed %v\n", err)
			return
		}
		err = sdk.loadAccount()
		if err != nil {
			fmt.Printf("load account failed. Is ~/.iwallet/<accountName>_ed25519 exists?\n")
			return
		}
		err = sdk.CreateNewAccount(newName, newKp, 10, 600, 0)
		if err != nil {
			fmt.Printf("create new account error %v\n", err)
			return
		}
		err = sdk.saveAccount(newName, newKp)
		if err != nil {
			fmt.Printf("saveAccount failed %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
}
