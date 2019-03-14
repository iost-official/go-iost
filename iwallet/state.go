package iwallet

import (
	"fmt"
	"strings"

	"github.com/iost-official/go-iost/sdk"
	"github.com/spf13/cobra"
)

// stateCmd prints the state of blockchain.
var stateCmd = &cobra.Command{
	Use:     "state",
	Short:   "Get blockchain and node state",
	Long:    `Get blockchain and node state`,
	Example: `  iwallet state`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := iwalletSDK.Connect(); err != nil {
			return err
		}
		defer iwalletSDK.CloseConn()
		n, err := iwalletSDK.GetNodeInfo()
		if err != nil {
			return fmt.Errorf("cannot get node info: %v", err)
		}
		fmt.Print(strings.TrimRight(sdk.MarshalTextString(n), "}"))
		c, err := iwalletSDK.GetChainInfo()
		if err != nil {
			return fmt.Errorf("cannot get chain info: %v", err)
		}
		fmt.Println(strings.Replace(sdk.MarshalTextString(c), "{\n", "", 1))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
}
