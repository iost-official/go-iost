package iwallet

import (
	"encoding/json"
	"fmt"

	rpcpb "github.com/iost-official/go-iost/rpc/pb"
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
		c, err := iwalletSDK.GetChainInfo()
		if err != nil {
			return fmt.Errorf("cannot get chain info: %v", err)
		}

		s := &struct {
			*rpcpb.NodeInfoResponse
			*rpcpb.ChainInfoResponse
		}{n, c}

		r, err := json.MarshalIndent(s, "", "    ")
		if err != nil {
			fmt.Println("json.Marshal error: " + err.Error())
		} else {
			fmt.Println(string(r))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
}
