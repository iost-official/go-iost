package iwallet

import (
	"fmt"

	"github.com/spf13/cobra"
)

var memo string

var transferCmd = &cobra.Command{
	Use:     "transfer receiver amount",
	Aliases: []string{"trans"},
	Short:   "Transfer IOST",
	Long:    `Transfer IOST`,
	Example: `  iwallet transfer test1 100 --account test0
  iwallet transfer test1 100 --account test0 --memo "just for test :D\n中文测试\n😏"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "receiver", "amount"); err != nil {
			return err
		}
		if err := checkFloat(cmd, args[1], "amount"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set account since making actions needs accountName.
		err := initAccountForSDK(iwalletSDK)
		if err != nil {
			return err
		}
		if accountName == "" {
			return fmt.Errorf("invalid account name")
		}
		return processMethod("token.iost", "transfer", "iost", accountName, args[0], args[1], memo)
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
	transferCmd.Flags().StringVarP(&memo, "memo", "", "", "memo of transfer")
}
