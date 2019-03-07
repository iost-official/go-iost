package iwallet

import (
	"fmt"
	"github.com/iost-official/go-iost/sdk"

	"github.com/spf13/cobra"
)

// accountInfoCmd represents the balance command.
var accountInfoCmd = &cobra.Command{
	Use:     "balance accountName",
	Short:   "Check the information of a specified account",
	Long:    `Check the information of a specified account`,
	Example: `  iwallet balance test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "accountName"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		info, err := iwalletSDK.GetAccountInfo(id)
		if err != nil {
			return err
		}
		fmt.Println(sdk.MarshalTextString(info))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountInfoCmd)
}
