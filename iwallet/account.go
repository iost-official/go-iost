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
	"path/filepath"

	"os"
	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/crypto"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "KeyPair manage",
	Long:  `Manage account of local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case nickName != "no":
			if strings.ContainsAny(nickName, `?*:|/\"`) || len(nickName) > 16 {
				fmt.Println("invalid nick name")
			}
			algo := getSignAlgo(signAlgo)
			ac, _ := account.NewKeyPair(nil, algo)
			if !filepath.IsAbs(kvPath) {
				kvPath, _ = filepath.Abs(kvPath)
			}
			if err := os.MkdirAll(kvPath, 0700); err != nil {
				panic(err)
			}
			fileName := kvPath + "/" + nickName
			if algo == crypto.Ed25519 {
				fileName += "_ed25519"
			}
			if algo == crypto.Secp256k1 {
				fileName += "_secp256k1"
			}
			pubfile, err := os.Create(fileName + ".pub")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer pubfile.Close()

			_, err = pubfile.WriteString(saveBytes(ac.Pubkey))
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			secFile, err := os.Create(fileName)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer secFile.Close()

			_, err = secFile.WriteString(saveBytes(ac.Seckey))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			idFileName := fileName + ".id"
			idFile, err := os.Create(idFileName)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer idFile.Close()
			id := account.GetIDByPubkey(ac.Pubkey)
			_, err = idFile.WriteString(id)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("the iost account ID is:")
			fmt.Println(id)
			fmt.Println("your account id is saved at:")
			fmt.Println(idFileName)
			fmt.Println("your account private key is saved at:")
			fmt.Println(fileName)
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

	accountCmd.Flags().StringVarP(&nickName, "name", "n", "id", "Create new account, using input as nickname")
	accountCmd.Flags().StringVarP(&kvPath, "path", "p", home+"/.iwallet", "Set path of key pair file")
	accountCmd.Flags().StringVarP(&signAlgo, "signAlgo", "a", "ed25519", "Sign algorithm")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
