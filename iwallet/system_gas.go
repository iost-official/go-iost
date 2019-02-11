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
	"strconv"

	"github.com/spf13/cobra"
)

var gas_user string

var pledgeCmd = &cobra.Command{
	Use:     "gas-pledge amount",
	Aliases: []string{"pledge"},
	Short:   "Pledge IOST to obtain gas",
	Long:    `Pledge IOST to obtain gas`,
	Example: `  iwallet sys pledge 100 --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "amount"); err != nil {
			return err
		}
		if err := checkFloat(cmd, args[0], "amount"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if gas_user == "" {
			gas_user = sdk.accountName
		}
		amount, _ := strconv.ParseFloat(args[0], 64)
		return sendAction("gas.iost", "pledge", sdk.accountName, gas_user, amount)
	},
}

var unpledgeCmd = &cobra.Command{
	Use:     "gas-unpledge amount",
	Aliases: []string{"unpledge"},
	Short:   "Undo pledge",
	Long:    `Undo pledge and get back the IOST pledged ealier`,
	Example: `  iwallet sys unpledge 100 --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "amount"); err != nil {
			return err
		}
		if err := checkFloat(cmd, args[0], "amount"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if gas_user == "" {
			gas_user = sdk.accountName
		}
		amount, _ := strconv.ParseFloat(args[0], 64)
		return sendAction("gas.iost", "unpledge", sdk.accountName, gas_user, amount)
	},
}

func init() {
	systemCmd.AddCommand(pledgeCmd)
	pledgeCmd.Flags().StringVarP(&gas_user, "gas_user", "", "", "gas user that pledge IOST for (default is pledger himself/herself)")
	systemCmd.AddCommand(unpledgeCmd)
	unpledgeCmd.Flags().StringVarP(&gas_user, "gas_user", "", "", "gas user that earlier pledge for (default is pledger himself/herself)")
}
