package iwallet

import (
	"fmt"

	rpcpb "github.com/iost-official/go-iost/v3/rpc/pb"
	"github.com/iost-official/go-iost/v3/sdk"
	"github.com/spf13/cobra"
)

// callCmd represents the call command that call a contract with given actions.
var callCmd = &cobra.Command{
	Use:   "call [ACTION]...",
	Short: "Call the method in contracts",
	Long: `Call the method in contracts
	Would accept arguments as call actions or load transaction request directly from given file (which could be generated by "save" command).
	An ACTION is a group of 3 arguments: contract name, function name, method parameters.
	The method parameters should be a string with format '["arg0","arg1",...]'.`,
	Example: `  iwallet call "token.iost" "transfer" '["iost","user0001","user0002","123.45",""]' --account test0
  iwallet call "token.iost" "transfer" '["iost","user0001","user0002","123.45",""]' --output tx.json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if outputTxFile == "" {
			return checkAccount(cmd)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		argc := len(args)
		if argc%3 != 0 {
			return fmt.Errorf(`number of args should be a multiplier of 3`)
		}
		var actions = make([]*rpcpb.Action, 0)
		for i := 0; i < len(args); i += 3 {
			v, err := formatContractArgs(args[i+2])
			if err != nil {
				return err
			}
			act := sdk.NewAction(args[i], args[i+1], v)
			actions = append(actions, act)
		}
		tx, err := createTxFromActions(actions)
		if err != nil {
			return err
		}
		_, err = processTx(tx)
		return err
	},
}

func init() {
	rootCmd.AddCommand(callCmd)
}
