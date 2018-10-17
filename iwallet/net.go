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
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/iost-official/go-iost/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// balanceCmd represents the balance command
var netCmd = &cobra.Command{
	Use:   "net",
	Short: "Get network id",
	Long:  `Get network id`,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := getNetID()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("netId:", b)
	},
}

func init() {
	rootCmd.AddCommand(netCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// balanceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// balanceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getNetID() (string, error) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()
	client := rpc.NewApisClient(conn)
	value, err := client.GetNodeInfo(context.Background(), &empty.Empty{})
	if err != nil {
		return "", err
	}

	return value.Network.ID, nil
}
