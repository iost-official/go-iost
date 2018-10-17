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
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/iost-official/go-iost/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// blockCmd represents the block command
var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "print block info, default find by block number",
	Long:  `print block info, default find by block number`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var i int64
		var err error
		if len(args) < 1 {
			fmt.Println(`Error: block num or hash not given`)
			return
		}

		switch method {
		case "num":
			i, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		case "hash":
		default:
			fmt.Println("please enter correct method arg")
			return
		}

		conn, err := grpc.Dial(server, grpc.WithInsecure())
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer conn.Close()
		client := rpc.NewApisClient(conn)
		var blockInfo *rpc.BlockInfo
		if method == "num" {
			blockInfo, err = client.GetBlockByNum(context.Background(), &rpc.BlockByNumReq{Num: i, Complete: complete})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			blockInfo, err = client.GetBlockByHash(context.Background(), &rpc.BlockByHashReq{Hash: args[0], Complete: complete})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
		blockInfoJSON, err := json.Marshal(blockInfo)
		fmt.Println(string(blockInfoJSON))
	},
}

var method string
var complete bool

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.Flags().StringVarP(&method, "method", "m", "num", "find by block num or hash")
	blockCmd.Flags().BoolVarP(&complete, "complete", "c", false, "indicate whether to fetch all the trxs in the block or not")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// blockCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// blockCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
