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
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var (
	createName       string
	viewAccounts     string
	ownerKey         string
	activeKey        string
	importAccount    string
	deleteMethod     string
	cryptoName       = []string{"ed25519", "secp256k1"}
	initialRAM       int64
	initialBalance   int64
	initialGasPledge int64
)

type acc struct {
	Name    string
	KeyPair *key
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		switch {
		case len(createName) != 0:
			return createAccount(createName)
		case len(viewAccounts) != 0:
			viewAccount(viewAccounts)
		case len(deleteMethod) != 0:
			delAccount(deleteMethod)
		case len(importAccount) != 0:
			importAcc(importAccount, args)
		}
		return nil

	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
	accountCmd.Flags().StringVarP(&createName, "create", "c", "", "create an account on blockchain")
	accountCmd.Flags().StringVarP(&viewAccounts, "accounts", "a", "", "view account by name or All")
	accountCmd.Flags().StringVarP(&ownerKey, "owner", "", "", "owner key")
	accountCmd.Flags().StringVarP(&activeKey, "active", "", "", "active key")
	accountCmd.Flags().StringVarP(&importAccount, "import", "i", "", "import an account, args[account_name account_private_key]")
	accountCmd.Flags().StringVarP(&deleteMethod, "delete", "", "", "delete an account")
	accountCmd.Flags().Int64VarP(&initialRAM, "initial_ram", "", 1024, "buy $initial_ram bytes ram for the new account")
	accountCmd.Flags().Int64VarP(&initialGasPledge, "initial_gas_pledge", "", 10, "pledge $initial_gas_pledge IOSTs for the new account")
	accountCmd.Flags().Int64VarP(&initialBalance, "initial_balance", "", 0, "transfer $initial_balance IOSTs to the new account")

}

func createAccount(name string) (err error) {
	var (
		autoKey    bool
		okey, akey string
		newKp      *account.KeyPair
	)

	newName := name
	if strings.ContainsAny(newName, `?*:|/\"`) || len(newName) > 16 {
		return fmt.Errorf("invalid account name")
	}

	if sdk.checkID(ownerKey) && sdk.checkID(activeKey) {
		okey, akey = ownerKey, activeKey
	} else {
		aLgo := sdk.getSignAlgo()
		newKp, err = account.NewKeyPair(nil, aLgo)
		if err != nil {
			return fmt.Errorf("create key pair failed %v", err)
		}
		okey, akey = newKp.ID, newKp.ID
		autoKey = true
	}

	err = sdk.loadAccount()
	if err != nil {
		return fmt.Errorf("load account failed. Is ~/.iwallet/<accountName>_ed25519 exists")
	}
	err = sdk.CreateNewAccount(newName, okey, akey, initialGasPledge, initialRAM, initialBalance)
	if err != nil {
		return fmt.Errorf("create new account error %v", err)
	}
	if autoKey {
		err = sdk.saveAccount(newName, newKp)
		if err != nil {
			return fmt.Errorf("saveAccount failed %v", err)
		}
	}

	fmt.Println("create account done")
	fmt.Println("the iost account ID is:", name)
	fmt.Println("owner permission key:", okey)
	fmt.Println("active permission key:", akey)

	return nil
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
				fsk, err := loadKey(f)
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
				var k key
				k.ID = keyPair.ID
				k.Algorithm = keyPair.Algorithm.String()
				k.Pubkey = common.Base58Encode(keyPair.Pubkey)
				k.Seckey = common.Base58Encode(keyPair.Seckey)
				al.Account = append(al.Account, &acc{name, &k})

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
			n := fmt.Sprintf("%s/%s_%s", dir, name, v)
			fsk, err := readFile(n)
			if err != nil {
				continue
			}
			keyPair, err := account.NewKeyPair(loadBytes(string(fsk)), crypto.NewAlgorithm(v))
			if err != nil {
				fmt.Println("NewKeyPair error: ", err)
				continue
			}
			var k key
			k.ID = keyPair.ID
			k.Algorithm = keyPair.Algorithm.String()
			k.Pubkey = common.Base58Encode(keyPair.Pubkey)
			k.Seckey = common.Base58Encode(keyPair.Seckey)

			al.Account = append(al.Account, &acc{name, &k})
		}

		ret, err := json.MarshalIndent(al, "", "    ")
		if err != nil {
			fmt.Println("json.Marshal error: ", err)
			return
		}
		fmt.Println(string(ret))
	}
}

func importAcc(name string, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: private key not given")
		return
	}
	keyPair, err := account.NewKeyPair(loadBytes(args[0]), sdk.getSignAlgo())
	if err != nil {
		fmt.Println("private key error: ", err)
		return
	}
	err = sdk.saveAccount(name, keyPair)
	if err != nil {
		fmt.Printf("saveAccount failed %v\n", err)
	}

	fmt.Println("import account done")
	fmt.Println("the iost account ID is:", name)
	fmt.Println("active permission:", keyPair.ID)
}

func delAccount(name string) error {
	dir, err := sdk.getAccountDir()
	if err != nil {
		return fmt.Errorf("getAccountDir error: %v", err)
	}
	for _, v := range cryptoName {
		n := fmt.Sprintf("%s/%s_%s", dir, name, v)
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
