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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/spf13/cobra"
)

var method string
var complete bool

// blockCmd represents the block command
var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "print block info, default find by block number",
	Long:  `print block info, default find by block number`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(`Error: block num or hash not given`)
			return
		}
		var blockInfo *rpcpb.BlockResponse
		var num int64
		var err error
		if method == "num" {
			num, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Printf("invalid block number %v\n", err)
				return
			}
			blockInfo, err = sdk.getGetBlockByNum(num, complete)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		} else if method == "hash" {
			blockInfo, err = sdk.getGetBlockByHash(args[0], complete)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		} else {
			fmt.Println("please enter correct method arg")
			return
		}
		ret, err := json.MarshalIndent(blockInfo, "", "    ")
		if err != nil {
			fmt.Printf("error %v\n", err)
			return
		}

		fmt.Println(string(ret))
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.Flags().StringVarP(&method, "method", "m", "num", "find by block num or hash")
	blockCmd.Flags().BoolVarP(&complete, "complete", "c", false, "indicate whether to fetch all the trxs in the block or not")
}
