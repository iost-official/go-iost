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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
)

func saveAccount(name string, kp *account.KeyPair) {
	if !filepath.IsAbs(kvPath) {
		kvPath, _ = filepath.Abs(kvPath)
	}

	if err := os.MkdirAll(kvPath, 0700); err != nil {
		panic(err)
	}
	fileName := kvPath + "/" + name
	if kp.Algorithm == crypto.Ed25519 {
		fileName += "_ed25519"
	}
	if kp.Algorithm == crypto.Secp256k1 {
		fileName += "_secp256k1"
	}

	pubfile, err := os.Create(fileName + ".pub")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer pubfile.Close()

	_, err = pubfile.WriteString(saveBytes(kp.Pubkey))
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

	_, err = secFile.WriteString(saveBytes(kp.Seckey))
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
	id := account.GetIDByPubkey(kp.Pubkey)
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
}

func createNewAccount(creatorID string, creatorKp *account.KeyPair, newID string, newKp *account.KeyPair) {
	var acts []*tx.Action
	acts = append(acts, tx.NewAction("iost.auth", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, newID, newKp.ID, newKp.ID)))
	acts = append(acts, tx.NewAction("iost.gas", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, creatorID, newID, 100)))
	acts = append(acts, tx.NewAction("iost.ram", "buy", fmt.Sprintf(`["%v", "%v", %v]`, creatorID, newID, 100)))
	trx := tx.NewTx(acts, make([]string, len(signers)), 10000, 100, time.Now().Add(time.Second*time.Duration(5)).UnixNano(), 0)
	stx, err := tx.SignTx(trx, creatorID, []*account.KeyPair{creatorKp})
	if err != nil {
		panic(err)
	}
	var txHash string
	txHash, err = sendTx(stx)
	if err != nil {
		panic(err)
	}
	fmt.Println("iost node:receive your tx!")
	fmt.Println("the transaction hash is:", txHash)
	if checkResult {
		checkTransaction(txHash)
	}
	info, err := GetAccountInfo(server, newID, useLongestChain)
	if err != nil {
		panic(err)
	}
	fmt.Println(info)
}

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "KeyPair manage",
	Long:  `Manage account of local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		if accountName == "" {
			panic("you must provide account name")
		}
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		kpPath := fmt.Sprintf("%s/.iwallet/%s_ed25519", home, accountName)
		fsk, err := readFile(kpPath)
		if err != nil {
			fmt.Println("Read file failed: ", err.Error())
			return
		}

		keyPair, err := account.NewKeyPair(loadBytes(string(fsk)), getSignAlgo(signAlgo))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		newName := args[0]
		if strings.ContainsAny(newName, `?*:|/\"`) || len(newName) > 16 {
			panic("invalid nick name")
		}
		algo := getSignAlgo(signAlgo)
		newKp, err := account.NewKeyPair(nil, algo)
		if err != nil {
			panic(err)
		}
		createNewAccount(accountName, keyPair, newName, newKp)
		saveAccount(newName, newKp)
	},
}

var kvPath string
var accountName string

func init() {
	rootCmd.AddCommand(accountCmd)

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	accountCmd.Flags().StringVarP(&accountName, "account", "", "id", "Create new account, using input as nickname")
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
