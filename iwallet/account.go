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
	"encoding/json"
	"fmt"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/crypto"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var (
	createName   string
	viewAccounts string
	deleteMethod string
	cryptoName   = []string{"ed25519", "secp256k1"}
)

type acc struct {
	Name    string
	KeyPair *account.KeyPair
}
type accounts struct {
	Dir     string
	Account []*acc
}

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "KeyPair manage",
	Long:  `Manage account of local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		//if len(createName) != 0 {
		//
		//}
		switch {
		case len(createName) != 0:
			createAccount(createName)
		case len(viewAccounts) != 0:
			viewAccount(viewAccounts)
		case len(deleteMethod) != 0:
			delAccount(deleteMethod)
		}

	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
	accountCmd.Flags().StringVarP(&createName, "create", "c", "", "create an account on blockchain")
	accountCmd.Flags().StringVarP(&viewAccounts, "accounts", "a", "", "view account by name or All")
	accountCmd.Flags().StringVarP(&deleteMethod, "delete", "", "", "delete an account")
}

func createAccount(name string) {
	newName := name
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
}

func viewAccount(name string) {

	dir, err := sdk.getAccountDir()
	if err != nil {
		fmt.Println("getAccountDir error: ", err)
		return
	}
	al := accounts{}
	al.Dir = dir
	if name == "All" {
		for _, v := range cryptoName {

			files, err := getFilesAndDirs(dir, v)
			if err != nil {
				fmt.Println("getFilesAndDirs error: ", err)
				return
			}
			for _, f := range files {
				fsk, err := readFile(f)
				if err != nil {
					fmt.Println("read file failed: ", err)
					continue
				}
				keyPair, err := account.NewKeyPair(loadBytes(string(fsk)), crypto.NewAlgorithm(v))
				if err != nil {
					fmt.Println("NewKeyPair error: ", err)
					continue
				}
				name, err := getFileName(f, "_"+v)
				if err != nil {
					fmt.Println("getFileName error: ", err)
					continue
				}
				al.Account = append(al.Account, &acc{name, keyPair})

			}
			if len(files) != 0 {
				ret, err := json.MarshalIndent(al, "", "    ")
				if err != nil {
					fmt.Println("json.Marshal error: ", err)
					return
				}
				fmt.Println(string(ret))
				al = accounts{}
			}
		}
	} else {
		for _, v := range cryptoName {
			n := dir + "/" + name + "_" + v
			fsk, err := readFile(n)
			if err != nil {
				continue
			}
			keyPair, err := account.NewKeyPair(loadBytes(string(fsk)), crypto.NewAlgorithm(v))
			if err != nil {
				fmt.Println("NewKeyPair error: ", err)
				continue
			}
			al.Account = append(al.Account, &acc{name, keyPair})
		}

		ret, err := json.MarshalIndent(al, "", "    ")
		if err != nil {
			fmt.Println("json.Marshal error: ", err)
			return
		}
		fmt.Println(string(ret))
	}
}

func delAccount(name string) error {
	dir, err := sdk.getAccountDir()
	if err != nil {
		return fmt.Errorf("getAccountDir error: %v", err)
	}
	for _, v := range cryptoName {
		n := dir + "/" + name + "_" + v
		os.Remove(n)
		os.Remove(n + ".id")
		os.Remove(n + ".pub")
	}
	fmt.Printf("delete %v success\n", name)
	return nil
}

func getFileName(file string, suf string) (string, error) {
	f := file
	startIndex := strings.LastIndex(f, "/")
	if startIndex == -1 {
		return "", fmt.Errorf("file name error")
	}

	lastIndex := strings.LastIndex(f, suf)
	if lastIndex == -1 {
		return "", fmt.Errorf("file name error")
	}

	return f[startIndex+1 : lastIndex], nil
}

func getFilesAndDirs(dirPth string, suf string) (files []string, err error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if !fi.IsDir() {
			ok := strings.HasSuffix(fi.Name(), suf)
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	return files, nil
}
