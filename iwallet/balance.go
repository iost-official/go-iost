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

	"context"

	"github.com/iost-official/go-iost/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "check balance of specified account",
	Long:  `check balance of specified account`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("please enter the account ID")
			return
		}
		//do some check for arg[0] here
		id := args[0]
		info, err := GetAccountInfo(server, id, useLongestChain)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(info)
	},
}

var useLongestChain bool

func init() {
	rootCmd.AddCommand(balanceCmd)
	balanceCmd.Flags().BoolVarP(&useLongestChain, "use_longest", "l", false, "get balance on longest chain")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// balanceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// balanceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// GetAccountInfo return account info
func GetAccountInfo(server string, id string, useLongestChain bool) (*rpc.GetAccountRes, error) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpc.NewApisClient(conn)
	req := rpc.GetAccountReq{ID: id}
	if useLongestChain {
		req.UseLongestChain = true
	}
	value, err := client.GetAccountInfo(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return value, nil
}
