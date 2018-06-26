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

	"github.com/iost-official/prototype/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// valueCmd represents the value command
var valueCmd = &cobra.Command{
	Use:   "value",
	Short: "check value of a specified key",
	Long:  `check value of a specified key, from a iserver node`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		conn, err := grpc.Dial(server, grpc.WithInsecure())
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer conn.Close()
		client := rpc.NewCliClient(conn)
		for _, arg := range args {
			key := *prefix + arg
			st, err := client.GetState(context.Background(), &rpc.Key{S: key})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(st.Sv[1:])
		}

	},
}

var prefix *string

func init() {
	rootCmd.AddCommand(valueCmd)

	prefix = valueCmd.Flags().StringP("prefix", "p", "", "Set prefix of key")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// valueCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// valueCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
