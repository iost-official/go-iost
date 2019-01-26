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
	"os/exec"
	"strings"

	"github.com/iost-official/go-iost/iwallet/contract"
	"github.com/spf13/cobra"
)

// Generate ABI file.
func generateABI(codePath string) (string, error) {
	contractToRun := fmt.Sprintf(`
	const fs = require('fs');
	%s;
	const processContract = module.exports;
	processContract("%s");
	`, contract.CompiledContract, codePath)

	cmd := exec.Command("node", "-")
	cmd.Stdin = strings.NewReader(contractToRun)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to compile, error: %v\n", err)
		fmt.Println(string(output))
		fmt.Printf("Please make sure node.js has been installed\n")
		return "", err
	}

	return codePath + ".abi", nil
}

// compileCmd represents the compile command.
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Generate contract abi",
	Long: `Generate abi from contract javascript code
	Example:
		iwallet compile ./example.js
	`,

	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) < 1 {
			return fmt.Errorf("source code file not given")
		}
		codePath := args[0]
		abiPath, err := generateABI(codePath)
		if err != nil {
			return fmt.Errorf("failed to generate abi: %v", err)
		}
		fmt.Printf("Successfully generated abi file as: %v\n", abiPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(compileCmd)
}
