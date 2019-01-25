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
	"strconv"

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/spf13/cobra"
)

var byNum bool
var complete bool

// blockCmd represents the block command.
var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Print block info",
	Long:  `Print block info by block number or hash`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("Please enter the block number or hash")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var blockInfo *rpcpb.BlockResponse
		var num int64
		if byNum {
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
		} else {
			blockInfo, err = sdk.getGetBlockByHash(args[0], complete)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		}
		fmt.Println(marshalTextString(blockInfo))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.Flags().BoolVarP(&byNum, "by_num", "n", true, "find by block num or set false to find by hash")
	blockCmd.Flags().BoolVarP(&complete, "complete", "c", false, "indicate whether to fetch all the transactions in the block or not")
}
