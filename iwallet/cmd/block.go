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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/iost-official/prototype/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// blockCmd represents the block command
var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "print block info, default find by block number reversed",
	Long:  `"print block info, default find by block number reversed`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case *byHash:
			fallthrough
		case *byNumber:
			fmt.Println("not support yet")
			return
		case *byNumberR:
			fallthrough
		default:
		}

		i, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		conn, err := grpc.Dial(server, grpc.WithInsecure())
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer conn.Close()
		client := rpc.NewCliClient(conn)
		blockInfo, err := client.GetBlock(context.Background(), &rpc.BlockKey{Layer: int64(i)})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		blockInfoJson, err := json.Marshal(blockInfo)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(string(blockInfoJson))
	},
}

var byNumberR *bool
var byHash *bool
var byNumber *bool

func init() {
	rootCmd.AddCommand(blockCmd)

	byNumber = blockCmd.Flags().Bool("by-number-reverse", false, "find by layer")
	byHash = blockCmd.Flags().Bool("by-hash", false, "find by block head hash")
	byNumberR = blockCmd.Flags().Bool("by-number", false, "find by block head hash")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// blockCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// blockCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
