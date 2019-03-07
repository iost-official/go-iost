package iwallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iost-official/go-iost/sdk"
)

// transactionCmd represents the transaction command.
var transactionCmd = &cobra.Command{
	Use:     "transaction transactionHash",
	Aliases: []string{"tx"},
	Short:   "Find transactions",
	Long:    `Find transaction by transaction hash`,
	Example: `  iwallet transaction 7MDfKBeZToQnnfNHD58cbZ7o4Y2AktKLmiEg776HLPBT
  iwallet tx 7MDfKBeZToQnnfNHD58cbZ7o4Y2AktKLmiEg776HLPBT`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "transactionHash"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		txRaw, err := iwalletSDK.GetTxByHash(args[0])
		if err != nil {
			return err
		}
		fmt.Println(sdk.MarshalTextString(txRaw))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(transactionCmd)
}
