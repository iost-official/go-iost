package iwallet

import (
	"fmt"

	"github.com/iost-official/go-iost/v3/sdk"
	"github.com/spf13/cobra"
)

// receiptCmd represents the receipt command.
var receiptCmd = &cobra.Command{
	Use:     "receipt transactionHash",
	Short:   "Find receipt",
	Long:    `Find receipt by transaction hash`,
	Example: `  iwallet receipt 7MDfKBeZToQnnfNHD58cbZ7o4Y2AktKLmiEg776HLPBT`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "transactionHash"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		txReceipt, err := iwalletSDK.GetTxReceiptByTxHash(args[0])
		if err != nil {
			return err
		}
		fmt.Println(sdk.MarshalTextString(txReceipt))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(receiptCmd)
}
