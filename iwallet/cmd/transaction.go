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
	"context"
	"fmt"

	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// transactionCmd represents the transaction command
var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "find transactions",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if publisher == nil || nonce == nil {
			fmt.Println("input publisher and nonce")
			return
		}

		conn, err := grpc.Dial(server, grpc.WithInsecure())
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer conn.Close()
		client := rpc.NewCliClient(conn)
		txRaw, err := client.GetTransaction(context.Background(), &rpc.TransactionKey{Publisher: *publisher, Nonce: int64(*nonce)})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("tx raw:", txRaw.Tx)
		var txx tx.Tx
		err = txx.Decode(txRaw.Tx)
		if err != nil {
			fmt.Println(err.Error())
		}
		PrintTx(txx)
	},
}

var publisher *string
var nonce *int

func init() {
	rootCmd.AddCommand(transactionCmd)

	publisher = transactionCmd.Flags().StringP("publisher", "p", "", "find with publisher")
	nonce = transactionCmd.Flags().IntP("nonce", "n", -1, "find with nonce")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transactionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transactionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
