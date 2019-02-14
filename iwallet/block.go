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
	"github.com/iost-official/go-iost/sdk"
	"strconv"

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/spf13/cobra"
)

var method string
var complete bool

var methodMap = map[string]func(string) (*rpcpb.BlockResponse, error){
	"num": func(arg string) (*rpcpb.BlockResponse, error) {
		num, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			err = fmt.Errorf("invalid block number: %v", err)
			return nil, err
		}
		return iwalletSDK.GetBlockByNum(num, complete)
	},
	"hash": func(arg string) (*rpcpb.BlockResponse, error) {
		return iwalletSDK.GetBlockByHash(arg, complete)
	},
}

// blockCmd represents the block command.
var blockCmd = &cobra.Command{
	Use:   "block blockNum|blockHash",
	Short: "Print block info",
	Long:  `Print block info by block number or hash`,
	Example: `  iwallet block 0
  iwallet block 5dEgmyMURGfe7GxvTLajmaLXTkcqs5JwiJ2C2DE5VvVX -m hash`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmd.Usage()
			return fmt.Errorf("please enter the block number or hash")
		}
		_, ok := methodMap[method]
		if !ok {
			cmd.Usage()
			return fmt.Errorf("wrong method: %v", method)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blockInfo, err := methodMap[method](args[0])
		if err != nil {
			return err
		}
		fmt.Println(sdk.MarshalTextString(blockInfo))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.Flags().StringVarP(&method, "method", "m", "num", `find by block num (set as "num") or hash (set as "hash")`)
	blockCmd.Flags().BoolVarP(&complete, "complete", "c", false, "indicate whether to fetch all the transactions in the block or not")
}
