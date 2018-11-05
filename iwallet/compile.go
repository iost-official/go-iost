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

	"time"

	"go/build"
	"os/exec"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// generate ABI file
func generateABI(codePath string) string {
	var contractPath = ""
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	if _, err := os.Stat(home + "/.iwallet/contract_path"); err == nil {
		fd, err := readFile(home + "/.iwallet/contract_path")
		if err != nil {
			fmt.Println("Read ", home+"/.iwallet/contract_path file failed: ", err.Error())
			return ""
		}
		contractPath = string(fd)
	}

	if contractPath == "" {
		contractPath = gopath + "/src/github.com/iost-official/go-iost/iwallet/contract"
	}
	fmt.Println("contractPath: ", contractPath)
	cmd := exec.Command("node", contractPath+"/contract.js", codePath)

	err := cmd.Run()
	if err != nil {
		fmt.Println("run ", "node", contractPath, "/contract.js ", codePath, " Failed,error: ", err.Error())
		return ""
	}

	return codePath + ".abi"
}

// PublishContract converts contract js code to transaction. If 'send', also send it to chain.
func PublishContract(codePath string, abiPath string, conID string, acc *account.KeyPair, expiration int64,
	signers []string, gasLimit int64, gasPrice int64, update bool, updateID string, send bool) (stx *tx.Tx, txHash []byte, err error) {

	fd, err := readFile(codePath)
	if err != nil {
		fmt.Println("Read source code file failed: ", err.Error())
		return nil, nil, err
	}
	code := string(fd)

	fd, err = readFile(abiPath)
	if err != nil {
		fmt.Println("Read abi file failed: ", err.Error())
		return nil, nil, err
	}
	abi := string(fd)

	compiler := new(contract.Compiler)
	if compiler == nil {
		fmt.Println("gen compiler instance failed")
		return nil, nil, err
	}
	contract, err := compiler.Parse(conID, code, abi)
	if err != nil {
		fmt.Printf("gen contract error:%v\n", err)
		return nil, nil, err
	}

	methodName := "SetCode"
	data := `["` + contract.B64Encode() + `"]`
	if update {
		methodName = "UpdateCode"
		data = `["` + contract.B64Encode() + `", "` + updateID + `"]`
	}

	pubkeys := make([]string, len(signers))
	for i, accID := range signers {
		pubkeys[i] = accID
	}

	action := tx.NewAction("iost.system", methodName, data)
	trx := tx.NewTx([]*tx.Action{&action}, pubkeys, gasLimit, gasPrice, time.Now().Add(time.Second*time.Duration(expiration)).UnixNano(), delaySecond)
	if !send {
		return trx, nil, nil
	}
	stx, err = tx.SignTx(trx, acc.ID, []*account.KeyPair{acc})
	var hash []byte
	hash, err = sendTx(stx)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	fmt.Println("Sending tx to rpc server finished. The transaction hash is:", saveBytes(hash))
	return trx, hash, nil
}

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Generate tx",
	Long: `Generate a tx by a contract and an abi file
	example:iwallet compile -e 100 -l 10000 -p 1 --signers "IOSTdhsdhassdjaskd,IOSTskdjsakdjsk,..." ./example.js ./example.js.abi
	watch out:
		1.here signers means the account IDs who should sign this tx.for instance.if Alice transfers 100 to Bob in the constructor of the contract you want to deploy,then you should write Alice's account ID in --signers
		2.if you don't include --signers,iwallet consider this tx as a complete tx,So please use -k to indicate the file which contains your secret key to sign this tx.
		3.the secret key in the file indicate by -k is to sign the tx,meanwhile the account IDs in --signers are just to indicate this tx needs the authorization of these people.And they will generate their signatures later in the command 'iwallet sign'.
	`,

	Run: func(cmd *cobra.Command, args []string) {
		if resetContractPath {
			err := os.Remove(home + "/.iwallet/contract_path")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("Successfully reset contract path", setContractPath)
			return
		}

		//set contract path and save it to home/.iwallet/contract_path
		if setContractPath != "" {
			contractPathFile, err := os.Create(home + "/.iwallet/contract_path")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer contractPathFile.Close()

			_, err = contractPathFile.WriteString(setContractPath)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("Successfully set contract path to: ", setContractPath)
			return
		}

		if len(args) < 1 {
			fmt.Println(`Error: source code file not given`)
			return
		}
		codePath := args[0]

		var abiPath string

		if genABI {
			abiPath = generateABI(codePath)
			return
		} else if len(args) < 2 {
			fmt.Println(`Error: source code file or abi file not given`)
			return
		} else {
			abiPath = args[1]
			fmt.Println(args)
		}

		if abiPath == "" {
			fmt.Println("Failed to Gen ABI!")
			return
		}

		conID := ""
		if update {
			if len(args) < 3 {
				fmt.Println(`Error: contract id not given`)
				return
			}
			conID = args[2]
		}

		updateID := ""
		if update && len(args) >= 4 {
			updateID = args[3]
		}

		send := false
		var acc *account.KeyPair
		if len(signers) == 0 {
			fmt.Println("you don't indicate any signers,so this tx will be sent to the iostNode directly")
			fmt.Println("please ensure that the right secret key file path is given by parameter -k,or the secret key file path is ~/.iwallet/id_ed25519 by default,this file indicate the secret key to sign the tx")
			send = true
			fsk, err := readFile(kpPath)
			if err != nil {
				fmt.Println("Read file failed: ", err.Error())
				return
			}
			acc, err = account.NewKeyPair(loadBytes(string(fsk)), getSignAlgo(signAlgo))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
		trx, txHash, err := PublishContract(codePath, abiPath, conID, acc, expiration, signers, gasLimit, gasPrice, update, updateID, send)
		if err != nil {
			fmt.Println(err.Error())
		}
		if send {
			if checkResult {
				succ := checkTransaction(txHash)
				if succ {
					fmt.Println("The contract id is Contract" + saveBytes(txHash))
				}
			}
		} else {
			bytes := trx.Encode()
			if dest == "default" {
				dest = changeSuffix(args[0], ".sc")
			}
			err = saveTo(dest, bytes)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("the unsigned tx has been saved to %s\n", dest)
			fmt.Println("the account IDs of the signers are:", signers)
			fmt.Println("please inform them to sign this contract with the command 'iwallet sign' and send the generated signatures to you.by this step they give you the authorization,or this tx will fail to pass through the iost vm")
		}
	},
}

var dest string
var gasLimit int64
var gasPrice int64
var expiration int64
var delaySecond int64
var signers []string
var genABI bool
var update bool
var setContractPath string
var resetContractPath bool
var home string

func init() {
	rootCmd.AddCommand(compileCmd)
	var err error
	home, err = homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	compileCmd.Flags().Int64VarP(&gasLimit, "gaslimit", "l", 1000, "gasLimit for a transaction")
	compileCmd.Flags().Int64VarP(&gasPrice, "gasprice", "p", 1, "gasPrice for a transaction")
	compileCmd.Flags().Int64VarP(&expiration, "expiration", "e", 60*5, "expiration time for a transaction,for example,-e 60 means the tx will expire after 60 seconds from now on")
	compileCmd.Flags().Int64VarP(&delaySecond, "delaysecond", "d", 0, "delay time for a transaction,for example,-d 86400 means the tx will be excuted after 86400 seconds from when it's packed in block")
	compileCmd.Flags().StringSliceVarP(&signers, "signers", "n", []string{}, "signers who should sign this transaction")
	compileCmd.Flags().StringVarP(&kpPath, "key-path", "k", home+"/.iwallet/id_ed25519", "Set path of sec-key")
	compileCmd.Flags().StringVarP(&signAlgo, "signAlgo", "a", "ed25519", "Sign algorithm")
	compileCmd.Flags().BoolVarP(&genABI, "genABI", "g", false, "generate abi file")
	compileCmd.Flags().BoolVarP(&update, "update", "u", false, "update contract")
	compileCmd.Flags().StringVarP(&setContractPath, "setContractPath", "c", "", "set contract path, default is $GOPATH + /src/github.com/iost-official/go-iost/iwallet/contract")
	compileCmd.Flags().BoolVarP(&resetContractPath, "resetContractPath", "r", false, "clean contract path")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compi leCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
