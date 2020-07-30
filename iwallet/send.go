package iwallet

import (
	"github.com/spf13/cobra"

	rpcpb "github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
)

// sendCmd represents the send command that send a contract with given actions.
var sendCmd = &cobra.Command{
	Use:     "send txFile",
	Short:   "Send transaction onto blockchain by given json file",
	Long:    `Send transaction onto blockchain by given json file`,
	Example: `  iwallet send tx.json --account test0`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "txFile"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		tx := &rpcpb.TransactionRequest{}
		err := sdk.LoadProtoStructFromJSONFile(args[0], tx)
		if err != nil {
			return err
		}
		return sendTx(tx)
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
