package iwallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/iost-official/go-iost/v3/account"
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/sdk"
	"github.com/spf13/cobra"
)

type key struct {
	Algorithm string
	Pubkey    string
	Seckey    string
}

var privkey string
var encoding string

// keyCmd represents the keyPair command
var keyCmd = &cobra.Command{
	Use:     "key",
	Short:   "Create a key pair",
	Long:    `Create a key pair`,
	Example: `  iwallet key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyBytes := sdk.ParsePrivKey(privkey)
		if len(keyBytes) == 0 {
			keyBytes = nil
		}
		if signAlgo == "secp256k1" && len(keyBytes) == 64 {
			keyBytes = keyBytes[:32]
		}
		if signAlgo == "ed25519" && len(keyBytes) == 32 {
			// from seed
			keyBytes = ed25519.NewKeyFromSeed(keyBytes)
		}
		n, err := account.NewKeyPair(keyBytes, sdk.GetSignAlgoByName(signAlgo))
		if err != nil {
			return fmt.Errorf("failed to new key pair: %v", err)
		}

		var k key
		k.Algorithm = n.Algorithm.String()
		if encoding == "hex" {
			k.Pubkey = "0x" + hex.EncodeToString(n.Pubkey)
			k.Seckey = "0x" + hex.EncodeToString(n.Seckey)
		} else {
			// base58
			k.Pubkey = common.Base58Encode(n.Pubkey)
			k.Seckey = common.Base58Encode(n.Seckey)
		}

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
	keyCmd.Flags().StringVarP(&privkey, "privkey", "", "", "")
	keyCmd.Flags().StringVarP(&encoding, "encoding", "", "base58", "encoding used to show keys: hex|base58")
}
