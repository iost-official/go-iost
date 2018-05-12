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

	"github.com/iost-official/prototype/common"
	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(`invalid input, check
	iwallet publish -h`)
		} else if len(args) < 2 {
			fmt.Println(true) // TODO :签名之后发布
		}
		scf, err := os.Open(args[0])
		if err != nil {
			fmt.Printf("Error in File %v: %v\n", args[0], err.Error())
			return

		}
		defer scf.Close()
		sc, err := ioutil.ReadAll(scf)
		if err != nil {
			fmt.Println("Read error", err)
			return

		}
		for i, v := range args {
			if i == 0 {
				continue
			}
			sigf, err := os.Open(v)
			if err != nil {
				fmt.Printf("Error in File %v: %v\n", args[0], err.Error())
				return

			}
			defer scf.Close()
			sig, err := ioutil.ReadAll(sigf)
			if err != nil {
				fmt.Println("Read error", err)
				return
			}
			var sign common.Signature
			err = sign.Decode(sig)
			if err != nil {
				fmt.Println("Illegal sig file", err)
				return
			}
			if !common.VerifySignature(common.Sha256(sc), sign) {
				fmt.Printf("Sign %v error\n", v)
				return
			}
		}
		fmt.Println(true)

	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// publishCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
