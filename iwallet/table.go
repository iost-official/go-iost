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

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
	"github.com/spf13/cobra"
)

func getContractStorage(contract, key, field string) (*rpcpb.GetContractStorageResponse, error) {
	return iwalletSDK.GetContractStorage(&rpcpb.GetContractStorageRequest{
		Id:             contract,
		Key:            key,
		Field:          field,
		ByLongestChain: useLongestChain,
	})
}

var tableCmd = &cobra.Command{
	Use:   "table contract key [field]",
	Short: "Fetch stored info of given contract",
	Long:  `Fetch stored info of given contract`,
	Example: `  iwallet table vote_producer.iost currentProducerList
  iwallet table vote_producer.iost producerTable producer000`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "contract", "key"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var field string
		if len(args) > 2 {
			field = args[2]
		}
		response, err := getContractStorage(args[0], args[1], field)
		if err != nil {
			return err
		}
		fmt.Println(sdk.MarshalTextString(response))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tableCmd)
}
