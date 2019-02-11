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
	"io/ioutil"
	"os"
	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/spf13/cobra"
)

var signAlgos = []crypto.Algorithm{crypto.Ed25519, crypto.Secp256k1}

var (
	ownerKey         string
	activeKey        string
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

// accountCmd represents the account command.
var accountCmd = &cobra.Command{
	Use:     "account",
	Aliases: []string{"acc"},
	Short:   "KeyPair manager",
	Long:    `Manage account in local storage`,
}

var createCmd = &cobra.Command{
	Use:     "create accountName",
	Short:   "Create an account on blockchain",
	Long:    `Create an account on blockchain`,
	Example: `  iwallet account create test1 --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmd.Usage()
			return fmt.Errorf("please enter the account name")
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			autoKey    bool
			okey, akey string
			newKp      *account.KeyPair
			err        error
		)

		newName := args[0]
		if strings.ContainsAny(newName, `?*:|/\"`) || len(newName) < 5 || len(newName) > 11 {
			return fmt.Errorf("invalid account name")
		}

		if sdk.checkPubKey(ownerKey) && sdk.checkPubKey(activeKey) {
			okey, akey = ownerKey, activeKey
		} else {
			autoKey = true
			newKp, err = account.NewKeyPair(nil, sdk.GetSignAlgo())
			if err != nil {
				return fmt.Errorf("failed to create key pair: %v", err)
			}
			okey = newKp.ReadablePubkey()
			akey = okey
		}

		err = sdk.LoadAccount()
		if err != nil {
			return fmt.Errorf("failed to load account: %v", err)
		}
		_, err = sdk.CreateNewAccount(newName, okey, akey, initialGasPledge, initialRAM, initialBalance)
		if err != nil {
			return fmt.Errorf("create new account error: %v", err)
		}

		fmt.Println("The IOST account ID is:", newName)
		fmt.Println("Owner permission key:", okey)
		fmt.Println("Active permission key:", akey)

		if autoKey {
			err = sdk.SaveAccount(newName, newKp)
			if err != nil {
				return fmt.Errorf("failed to save account: %v", err)
			}
		}
		return nil
	},
}

var viewCmd = &cobra.Command{
	Use:   "view [accountName]",
	Short: "View account by name or omit to show all accounts",
	Long:  `View account by name or omit to show all accounts`,
	Example: `  iwallet account view test0
  iwallet account view`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := sdk.getAccountDir()
		if err != nil {
			return fmt.Errorf("failed to get account dir: %v", err)
		}
		a := accounts{}
		a.Dir = dir
		if len(args) < 1 {
			for _, algo := range signAlgos {
				files, err := getFilesAndDirs(dir, algo.String())
				if err != nil {
					return err
				}
				for _, f := range files {
					keyPair, err := loadKeyPair(f, algo)
					if err != nil {
						fmt.Println(err)
						continue
					}
					name, err := getAccountName(f, "_"+algo.String())
					if err != nil {
						return err
					}
					var k key
					k.Algorithm = keyPair.Algorithm.String()
					k.Pubkey = common.Base58Encode(keyPair.Pubkey)
					k.Seckey = common.Base58Encode(keyPair.Seckey)
					a.Account = append(a.Account, &acc{name, &k})
				}
			}
		} else {
			name := args[0]
			for _, algo := range signAlgos {
				f := fmt.Sprintf("%s/%s_%s", dir, name, algo.String())
				keyPair, err := loadKeyPair(f, algo)
				if err != nil {
					continue
				}
				var k key
				k.Algorithm = keyPair.Algorithm.String()
				k.Pubkey = common.Base58Encode(keyPair.Pubkey)
				k.Seckey = common.Base58Encode(keyPair.Seckey)
				a.Account = append(a.Account, &acc{name, &k})
			}
		}
		info, err := json.MarshalIndent(a, "", "    ")
		if err != nil {
			return err
		}
		fmt.Println(string(info))
		return nil
	},
}

var importCmd = &cobra.Command{
	Use:     "import accountName accountPrivateKey",
	Short:   "Import an account by name and private key",
	Long:    `Import an account by name and private key`,
	Example: `  iwallet account import test0 XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			cmd.Usage()
			return fmt.Errorf("please enter the account name and the private key")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		key := args[1]
		keyPair, err := account.NewKeyPair(common.Base58Decode(key), sdk.GetSignAlgo())
		if err != nil {
			return err
		}
		err = sdk.SaveAccount(name, keyPair)
		if err != nil {
			return fmt.Errorf("failed to save account: %v", err)
		}
		fmt.Println("The IOST account ID is:", name)
		fmt.Println("Active permission key:", keyPair.ReadablePubkey())
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete accountName",
	Aliases: []string{"del"},
	Short:   "Delete an account by name",
	Long:    `Delete an account by name`,
	Example: `  iwallet account delete test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmd.Usage()
			return fmt.Errorf("please enter the account name")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		dir, err := sdk.getAccountDir()
		if err != nil {
			return fmt.Errorf("failed to get account dir: %v", err)
		}
		found := false
		for _, algo := range signAlgos {
			f := fmt.Sprintf("%s/%s_%s", dir, name, algo.String())
			err = os.Remove(f)
			if err == nil {
				found = true
				fmt.Println("File", f, "has been removed.")
			}
			err = os.Remove(f + ".id")
			if err == nil {
				fmt.Println("File", f+".id", "has been removed.")
			}
			err = os.Remove(f + ".pub")
			if err == nil {
				fmt.Println("File", f+".pub", "has been removed.")
			}
		}
		if found {
			fmt.Println("Successfully deleted <", name, ">.")
		} else {
			fmt.Println("Account <", name, "> does not exist.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)

	accountCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&ownerKey, "owner", "", "", "owner key")
	createCmd.Flags().StringVarP(&activeKey, "active", "", "", "active key")
	createCmd.Flags().Int64VarP(&initialRAM, "initial_ram", "", 1024, "buy $initial_ram bytes ram for the new account")
	createCmd.Flags().Int64VarP(&initialGasPledge, "initial_gas_pledge", "", 10, "pledge $initial_gas_pledge IOSTs for the new account")
	createCmd.Flags().Int64VarP(&initialBalance, "initial_balance", "", 0, "transfer $initial_balance IOSTs to the new account")

	accountCmd.AddCommand(viewCmd)
	accountCmd.AddCommand(importCmd)
	accountCmd.AddCommand(deleteCmd)
}

func getAccountName(file string, suf string) (string, error) {
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
