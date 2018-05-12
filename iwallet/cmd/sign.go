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

	"github.com/iost-official/prototype/common"
	"github.com/spf13/cobra"
)

var Identity string
var kpPath string

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign to .sc file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		fpk, err := os.Open(kpPath + "/" + Identity + "_secp")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer fpk.Close()
		pubkey, err := ioutil.ReadAll(fpk)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		sig, err := common.Sign(common.Secp256k1, common.Sha256(fd), LoadBytes(string(pubkey)))
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if len(args) < 2 {
			Dist = args[0][:strings.LastIndex(args[0], ".")]
			Dist = Dist + ".sig"
		} else {
			Dist = args[1]
		}

		f, err := os.Create(Dist)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer f.Close()

		_, err = f.Write(sig.Encode())
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVarP(&Identity, "id", "i", "id", "Set language of contract, Support lua")
	signCmd.Flags().StringVarP(&kpPath, "key-path", "k", "test", "Set path of sec-key")

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// signCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// signCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
