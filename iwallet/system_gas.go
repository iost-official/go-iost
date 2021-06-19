package iwallet

import (
	"github.com/spf13/cobra"
)

var gasUser string

var pledgeCmd = &cobra.Command{
	Use:     "gas-pledge amount",
	Aliases: []string{"pledge"},
	Short:   "Pledge IOST to obtain gas",
	Long:    `Pledge IOST to obtain gas`,
	Example: `  iwallet sys pledge 100 --account test0
  iwallet sys pledge 100 --account test0 --gas_user test1`,
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
		if gasUser == "" {
			gasUser = accountName
		}
		return processMethod("gas.iost", "pledge", accountName, gasUser, args[0])
	},
}

var unpledgeCmd = &cobra.Command{
	Use:     "gas-unpledge amount",
	Aliases: []string{"unpledge"},
	Short:   "Undo pledge",
	Long:    `Undo pledge and get back the IOST pledged earlier`,
	Example: `  iwallet sys unpledge 100 --account test0
  iwallet sys pledge 100 --account test0 --gas_user test1`,
	Args: pledgeCmd.Args,
	RunE: func(cmd *cobra.Command, args []string) error {
		if gasUser == "" {
			gasUser = accountName
		}
		return processMethod("gas.iost", "unpledge", accountName, gasUser, args[0])
	},
}

func init() {
	systemCmd.AddCommand(pledgeCmd)
	pledgeCmd.Flags().StringVarP(&gasUser, "gas_user", "", "", "gas user that pledge IOST for (default is pledger himself/herself)")
	systemCmd.AddCommand(unpledgeCmd)
	unpledgeCmd.Flags().StringVarP(&gasUser, "gas_user", "", "", "gas user that earlier pledge for (default is pledger himself/herself)")
}
