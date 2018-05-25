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

	"context"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/rpc"
	"github.com/iost-official/prototype/vm"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "check balance of specified account",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var filePath string
		if len(args) < 1 {
			filePath = "~/.ssh/id_secp.pub"
		} else {
			filePath = args[0]
		}
		pubkey, err := ReadFile(filePath)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		pk := LoadBytes(string(pubkey))
		ia := vm.PubkeyToIOSTAccount(pk)
		b, err := CheckBalance(ia)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(filePath, ">", b, "iost")

	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// balanceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// balanceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func CheckBalance(ia vm.IOSTAccount) (float64, error) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	client := rpc.NewCliClient(conn)
	value, err := client.GetBalance(context.Background(), &rpc.Key{S: string(ia)})
	if err != nil {
		return 0, err
	}
	vv, err := state.ParseValue(value.Sv)
	if err != nil {
		return 0, err
	}
	return vv.(*state.VFloat).ToFloat64(), nil
}
