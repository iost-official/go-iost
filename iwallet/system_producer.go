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

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
	"github.com/spf13/cobra"
)

var location string
var url string
var networkID string
var isPartner bool
var publicKey string

var voteCmd = &cobra.Command{
	Use:     "producer-vote producerID amount",
	Aliases: []string{"vote"},
	Short:   "Vote a producer",
	Long:    `Vote a producer by given amount of IOSTs`,
	Example: `  iwallet sys vote producer000 1000000 --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "producerID", "amount"); err != nil {
			return err
		}
		if err := checkFloat(cmd, args[1], "amount"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "vote", accountName, args[0], args[1])
	},
}
var unvoteCmd = &cobra.Command{
	Use:     "producer-unvote producerID amount",
	Aliases: []string{"unvote"},
	Short:   "Unvote a producer",
	Long:    `Unvote a producer by given amount of IOSTs`,
	Example: `  iwallet sys unvote producer000 1000000 --account test0`,
	Args:    voteCmd.Args,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "unvote", accountName, args[0], args[1])
	},
}

var registerCmd = &cobra.Command{
	Use:     "producer-register publicKey",
	Aliases: []string{"register", "reg"},
	Short:   "Register as producer",
	Long:    `Register as producer`,
	Example: `  iwallet sys register XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX --account test0
  iwallet sys register XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX --account test1 --location PEK --url iost.io --net_id 123 --partner`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "publicKey"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "applyRegister", accountName, args[0], location, url, networkID, !isPartner)
	},
}
var unregisterCmd = &cobra.Command{
	Use:     "producer-unregister",
	Aliases: []string{"unregister", "unreg"},
	Short:   "Unregister from a producer",
	Long:    `Unregister from a producer`,
	Example: `  iwallet sys unregister --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "applyUnregister", accountName)
	},
}
var pcleanCmd = &cobra.Command{
	Use:     "producer-clean",
	Aliases: []string{"pclean"},
	Short:   "Clean producer info",
	Long:    `Clean producer info`,
	Example: `  iwallet sys pclean --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "unregister", accountName)
	},
}

var ploginCmd = &cobra.Command{
	Use:     "producer-login",
	Aliases: []string{"plogin"},
	Short:   "Producer login as online state",
	Long:    `Producer login as online state`,
	Example: `  iwallet sys plogin --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "logInProducer", accountName)
	},
}
var plogoutCmd = &cobra.Command{
	Use:     "producer-logout",
	Aliases: []string{"plogout"},
	Short:   "Producer logout as offline state",
	Long:    `Producer logout as offline state`,
	Example: `  iwallet sys plogout --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "logOutProducer", accountName)
	},
}

func getProducerVoteInfo(account string) (*rpcpb.GetProducerVoteInfoResponse, error) {
	return iwalletSDK.GetProducerVoteInfo(&rpcpb.GetProducerVoteInfoRequest{
		Account:        account,
		ByLongestChain: useLongestChain,
	})
}

var pinfoCmd = &cobra.Command{
	Use:     "producer-info producerID",
	Aliases: []string{"pinfo"},
	Short:   "Show producer info",
	Long:    `Show producer info`,
	Example: `  iwallet sys pinfo producer000`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "producerID"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := getProducerVoteInfo(args[0])
		if err != nil {
			return err
		}
		fmt.Println(sdk.MarshalTextString(info))
		return nil
	},
}

func getProducerList(key string) ([]string, error) {
	response, err := getContractStorage("vote_producer.iost", key, "")
	if err != nil {
		return nil, err
	}
	var list []string
	err = json.Unmarshal([]byte(response.Data), &list)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(list))
	for i, producerKey := range list {
		response, err := getContractStorage("vote_producer.iost", "producerKeyToId", producerKey)
		if err != nil {
			return nil, err
		}
		result[i] = response.Data
	}
	return result, nil
}

var plistCmd = &cobra.Command{
	Use:     "producer-list",
	Aliases: []string{"plist"},
	Short:   "Show current/pending producer list",
	Long:    `Show current/pending producer list`,
	Example: `  iwallet sys plist`,
	RunE: func(cmd *cobra.Command, args []string) error {
		currentList, err := getProducerList("currentProducerList")
		if err != nil {
			return err
		}
		fmt.Println("Current producer list:", currentList)
		pendingList, err := getProducerList("pendingProducerList")
		if err != nil {
			return err
		}
		fmt.Println("Pending producer list:", pendingList)
		return nil
	},
}

var pupdateCmd = &cobra.Command{
	Use:     "producer-update",
	Aliases: []string{"pupdate"},
	Short:   "Update producer info",
	Long:    `Update producer info`,
	Example: `  iwallet sys pupdate --account test0
  iwallet sys pupdate --account test1 --pubkey XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  iwallet sys pupdate --account test2 --location PEK --url iost.io --net_id 123`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := getProducerVoteInfo(accountName)
		if err != nil {
			return err
		}
		if publicKey == "" {
			publicKey = info.Pubkey
		}
		if location == "" {
			location = info.Loc
		}
		if url == "" {
			url = info.Url
		}
		if networkID == "" {
			networkID = info.NetId
		}
		return sendAction("vote_producer.iost", "updateProducer", accountName, publicKey, location, url, networkID)
	},
}

var predeemCmd = &cobra.Command{
	Use:     "producer-redeem [amount]",
	Aliases: []string{"predeem"},
	Short:   "Redeem the contribution value obtained by the block producing to IOST tokens",
	Long: `Redeem the contribution value obtained by the block producing to IOST tokens
	Omitting amount argument or zero amount will redeem all contribution value.`,
	Example: `  iwallet sys producer-redeem --account test0
  iwallet sys producer-redeem 10 --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if err := checkFloat(cmd, args[0], "amount"); err != nil {
				return err
			}
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		amount := "0"
		if len(args) > 0 {
			amount = args[0]
		}
		return sendAction("bonus.iost", "exchangeIOST", accountName, amount)
	},
}

var pwithdrawCmd = &cobra.Command{
	Use:     "producer-withdraw",
	Aliases: []string{"pwithdraw"},
	Short:   "Withdraw all voting reward for producer",
	Long:    `Withdraw all voting reward for producer`,
	Example: `  iwallet sys producer-withdraw --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "candidateWithdraw", accountName)
	},
}

var vwithdrawCmd = &cobra.Command{
	Use:     "voter-withdraw",
	Aliases: []string{"vwithdraw"},
	Short:   "Withdraw all voting reward for voter",
	Long:    `Withdraw all voting reward for voter`,
	Example: `  iwallet sys voter-withdraw --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendAction("vote_producer.iost", "voterWithdraw", accountName)
	},
}

func init() {
	systemCmd.AddCommand(voteCmd)
	systemCmd.AddCommand(unvoteCmd)

	systemCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringVarP(&location, "location", "", "", "location info")
	registerCmd.Flags().StringVarP(&url, "url", "", "", "url address")
	registerCmd.Flags().StringVarP(&networkID, "net_id", "", "", "network ID")
	registerCmd.Flags().BoolVarP(&isPartner, "partner", "", false, "if is partner instead of producer")
	systemCmd.AddCommand(unregisterCmd)
	systemCmd.AddCommand(pcleanCmd)

	systemCmd.AddCommand(ploginCmd)
	systemCmd.AddCommand(plogoutCmd)

	systemCmd.AddCommand(pinfoCmd)
	systemCmd.AddCommand(plistCmd)

	systemCmd.AddCommand(pupdateCmd)
	pupdateCmd.Flags().StringVarP(&publicKey, "pubkey", "", "", "publick key")
	pupdateCmd.Flags().StringVarP(&location, "location", "", "", "location info")
	pupdateCmd.Flags().StringVarP(&url, "url", "", "", "url address")
	pupdateCmd.Flags().StringVarP(&networkID, "net_id", "", "", "network ID")

	systemCmd.AddCommand(predeemCmd)
	systemCmd.AddCommand(pwithdrawCmd)
	systemCmd.AddCommand(vwithdrawCmd)
}
