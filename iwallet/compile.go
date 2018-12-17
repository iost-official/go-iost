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

	"go/build"
	"os/exec"

	"github.com/spf13/cobra"
)

var genABI bool
var update bool

//var setContractPath string
//var resetContractPath bool

// generate ABI file
func generateABI(codePath string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	contractPath := gopath + "/src/github.com/iost-official/go-iost/iwallet/contract"
	fmt.Println("node " + contractPath + "/contract.js " + codePath)
	cmd := exec.Command("node", contractPath+"/contract.js", codePath)
	err := cmd.Run()
	if err != nil {
		fmt.Println("run ", "node", contractPath, "/contract.js ", codePath, " Failed, error: ", err.Error())
		fmt.Println("Please make sure node.js has been installed")
		return "", err
	}

	return codePath + ".abi", nil
}

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Generate tx",
	Long: `Generate a tx by a contract and an abi file
	example:iwallet compile ./example.js ./example.js.abi
	`,

	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) < 1 {
			fmt.Println(`Error: source code file not given`)
			return
		}
		codePath := args[0]
		var abiPath string
		if genABI {
			abiPath, err := generateABI(codePath)
			if err != nil {
				return fmt.Errorf("failed to gen abi %v", err)
			}
			fmt.Printf("gen abi done. abi: %v\n", abiPath)
			return nil
		}

		if len(args) < 2 {
			fmt.Println(`Error: source code file or abi file not given`)
			return
		}

		abiPath = args[1]

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

		err = sdk.loadAccount()
		if err != nil {
			fmt.Printf("load account failed %v\n", err)
			return
		}
		_, txHash, err := sdk.PublishContract(codePath, abiPath, conID, update, updateID)
		if err != nil {
			fmt.Printf("create tx failed %v\n", err)
			return
		}
		if sdk.checkResult {
			succ := sdk.checkTransaction(txHash)
			if succ {
				fmt.Println("The contract id is Contract" + txHash)
			} else {
				return fmt.Errorf("check transaction failed")
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().BoolVarP(&genABI, "genABI", "g", false, "generate abi file")
	compileCmd.Flags().BoolVarP(&update, "update", "u", false, "update contract")
}
