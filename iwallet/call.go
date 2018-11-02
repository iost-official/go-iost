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
	"math"
	"os"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var checkResult bool
var checkResultDelay float32
var checkResultMaxRetry int32

func checkTransaction(txHash []byte) bool {
	// It may be better to to create a grpc client and reuse it. TODO later
	for i := int32(0); i < checkResultMaxRetry; i++ {
		time.Sleep(time.Duration(checkResultDelay*1000) * time.Millisecond)
		txReceipt, err := getTxReceiptByTxHash(server, saveBytes(txHash))
		if err != nil {
			fmt.Println("result not ready, please wait. Details: ", err)
			continue
		}
		if txReceipt == nil {
			fmt.Println("result not ready, please wait.")
			continue
		}
		if tx.StatusCode(txReceipt.Status.Code) != tx.Success {
			fmt.Println("exec tx failed: ", txReceipt.Status.Message)
			fmt.Println("full error information: ", txReceipt)
		} else {
			fmt.Println("exec tx done. gas used: ", txReceipt.GasUsage)
			return true
		}
		break
	}
	return false
}

// callCmd represents the compile command
var callCmd = &cobra.Command{
	Use:   "call",
	Short: "Call a method in some contract",
	Long: `Call a method in some contract
	the format of this command is:iwallet call contract_name0 function_name0 parameters0 contract_name1 function_name1 parameters1 ...
	(you can call more than one function in this command)
	the parameters is a string whose format is: ["arg0","arg1",...]
	example:./iwallet call -e 100 -l 10000 -p 1 "iost.system" "Transfer" '["IOST2g5LzaXkjAwpxCnCm29HK69wdbyRKbfG4BQQT7Yuqk57bgTFkY", "IOST25p7hEUu25YKEc8X9F8A7wXFJnMoWZtVVPVojM9LcCp2UEMhvg", 100]' -k ~/.iwallet/root_ed25519
	`,
	Run: func(cmd *cobra.Command, args []string) {
		argc := len(args)
		if argc%3 != 0 {
			fmt.Println(`Error: number of args should be a multiplier of 3`)
			return
		}
		var actions []*tx.Action = make([]*tx.Action, argc/3)
		for i := 0; i < len(args); i += 3 {
			// fixme use IOST as Measure Unit in iost.system Transfer, 1 IOST = 1e8
			if args[i] == "iost.system" && args[i+1] == "Transfer" {
				data, err := handleTransferData(args[i+2])
				if err != nil {
					fmt.Println("parse transfer amount failed. ", err)
					return
				}
				args[i+2] = data
			}
			act := tx.NewAction(args[i], args[i+1], args[i+2]) //check sth here
			actions[i] = &act
		}
		pubkeys := make([]string, len(signers))
		for i, accID := range signers {
			pubkeys[i] = accID
		}
		trx := tx.NewTx(actions, pubkeys, gasLimit, gasPrice, time.Now().Add(time.Second*time.Duration(expiration)).UnixNano(), delaySecond)
		if len(signers) == 0 {
			fmt.Println("you don't indicate any signers,so this tx will be sent to the iostNode directly")
			fmt.Println("please ensure that the right secret key file path is given by parameter -k,or the secret key file path is ~/.iwallet/id_ed25519 by default,this file indicate the secret key to sign the tx")
			fsk, err := readFile(kpPath)
			if err != nil {
				fmt.Println("Read file failed: ", err.Error())
				return
			}

			acc, err := account.NewKeyPair(loadBytes(string(fsk)), getSignAlgo(signAlgo))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			stx, err := tx.SignTx(trx, acc.ID, acc)
			var txHash []byte
			txHash, err = sendTx(stx)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("iost node:receive your tx!")
			fmt.Println("the transaction hash is:", saveBytes(txHash))
			if checkResult {
				checkTransaction(txHash)
			}
			return
		}

		bytes := trx.Encode()
		if dest == "default" {
			dest = changeSuffix(args[0], ".sc")
		}

		err := saveTo(dest, bytes)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("the unsigned tx has been saved to %s\n", dest)
		fmt.Println("the account IDs of the signers are:", signers)
		fmt.Println("please inform them to sign this contract with the command 'iwallet sign' and send the generated signatures to you.by this step they give you the authorization,or this tx will fail to pass through the iost vm")

	},
}

func init() {
	rootCmd.AddCommand(callCmd)

	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	callCmd.Flags().Int64VarP(&gasLimit, "gaslimit", "l", 1000, "gasLimit for a transaction")
	callCmd.Flags().Int64VarP(&gasPrice, "gasprice", "p", 1, "gasPrice for a transaction")
	callCmd.Flags().Int64VarP(&expiration, "expiration", "e", 60*5, "expiration time for a transaction,for example,-e 60 means the tx will expire after 60 seconds from now on")
	callCmd.Flags().StringSliceVarP(&signers, "signers", "n", []string{}, "signers who should sign this transaction")
	callCmd.Flags().StringVarP(&kpPath, "key-path", "k", home+"/.iwallet/id_ed25519", "Set path of sec-key")
	callCmd.Flags().StringVarP(&signAlgo, "signAlgo", "a", "ed25519", "Sign algorithm")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// callCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compi leCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func handleTransferData(data string) (string, error) {
	if strings.HasSuffix(data, ",]") {
		data = data[:len(data)-2] + "]"
	}
	js, err := simplejson.NewJson([]byte(data))
	if err != nil {
		return "", fmt.Errorf("error in data: %v", err)
	}

	arr, err := js.Array()
	if err != nil {
		ilog.Error(js.EncodePretty())
		return "", err
	}

	if len(arr) != 3 {
		return "", fmt.Errorf("Transfer need 3 arguments, got %v", len(arr))
	}
	if amount, err := js.GetIndex(2).Float64(); err == nil {
		if amount*1e8 > math.MaxInt64 {
			return "", fmt.Errorf("you can transfer more than %f iost", math.MaxInt64/1e8)
		}
		data = fmt.Sprintf(`["%v", "%v", %d]`, js.GetIndex(0).MustString(), js.GetIndex(1).MustString(), int64(amount*1e8))
	}
	return data, nil
}
