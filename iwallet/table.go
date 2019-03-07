package iwallet

import (
	"fmt"

	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
	"github.com/spf13/cobra"
)

func getContractStorage(contract, key, field string) (*rpcpb.GetContractStorageResponse, error) {
	return iwalletSDK.GetContractStorage(&rpcpb.GetContractStorageRequest{
		Id:             contract,
		Key:            key,
		Field:          field,
		ByLongestChain: useLongestChain,
	})
}

var tableCmd = &cobra.Command{
	Use:   "table contract key [field]",
	Short: "Fetch stored info of given contract",
	Long:  `Fetch stored info of given contract`,
	Example: `  iwallet table vote_producer.iost currentProducerList
  iwallet table vote_producer.iost producerTable producer000`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "contract", "key"); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var field string
		if len(args) > 2 {
			field = args[2]
		}
		response, err := getContractStorage(args[0], args[1], field)
		if err != nil {
			return err
		}
		fmt.Println(sdk.MarshalTextString(response))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tableCmd)
}
