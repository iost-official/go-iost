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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"strings"

	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	"github.com/spf13/cobra"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile contract files to smart contract",
	Long: `Compile contract files to smart contract. 
Useage : 

iwallet compile -l lua SRC`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(`Error: source file not given`)
			return
		}
		path := args[0]
		fi, err := os.Open(path)
		if err != nil {
			fmt.Println("Error: input file not found")
		}
		defer fi.Close()
		fd, err := ioutil.ReadAll(fi)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		rawCode := string(fd)

		var contract vm.Contract
		switch Language {
		case "lua":
			parser, _ := lua.NewDocCommentParser(rawCode)
			contract, err = parser.Parse()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		//		fmt.Printf(`Transaction :
		//Time: xx
		//Nonce: xx
		//Contract:
		//    Price: %v
		//    Gas limit: %v
		//Code:
		//----
		//%v
		//----
		//`, contract.Info().Price, contract.Info().GasLimit, contract.Code())

		mTx := tx.NewTx(int64(Nonce), contract)

		bytes := mTx.Encode()

		if Dist == "default" {
			Dist = args[0][:strings.LastIndex(args[0], ".")]
			Dist = Dist + ".sc"
		}

		f, err := os.Create(Dist)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer f.Close()

		_, err = f.Write(bytes)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	},
}

var Language string
var Dist string
var Nonce int

func init() {
	rootCmd.AddCommand(compileCmd)

	compileCmd.Flags().StringVarP(&Language, "language", "l", "lua", "Set language of contract, Support lua")
	compileCmd.Flags().StringVarP(&Dist, "dest", "d", "default", "Set destination of build file")
	compileCmd.Flags().IntVarP(&Nonce, "nonce", "n", 1, "Set Nonce of this Transaction")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
