package iwallet

import (
	"encoding/json"
	"fmt"

	"github.com/iost-official/go-iost/sdk"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/spf13/cobra"
)

type key struct {
	Algorithm string
	Pubkey    string
	Seckey    string
}

// keyCmd represents the keyPair command
var keyCmd = &cobra.Command{
	Use:     "key",
	Short:   "Create a key pair",
	Long:    `Create a key pair`,
	Example: `  iwallet key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := account.NewKeyPair(nil, sdk.GetSignAlgoByName(signAlgo))
		if err != nil {
			return fmt.Errorf("failed to new key pair: %v", err)
		}

		var k key
		k.Algorithm = n.Algorithm.String()
		k.Pubkey = common.Base58Encode(n.Pubkey)
		k.Seckey = common.Base58Encode(n.Seckey)

		ret, err := json.MarshalIndent(k, "", "    ")
		if err != nil {
			return fmt.Errorf("failed to marshal: %v", err)
		}
		fmt.Println(string(ret))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(keyCmd)
}
