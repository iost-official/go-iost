package iwallet

import (
	"fmt"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

func checkArgsNumber(cmd *cobra.Command, args []string, argNames ...string) error {
	if len(args) < len(argNames) {
		cmd.Usage()
		return fmt.Errorf("missing positional argument: %v", argNames[len(args):])
	}
	return nil
}

func checkAccount(cmd *cobra.Command) error {
	if accountName == "" {
		cmd.Usage()
		return fmt.Errorf("please provide the account name with flag --account")
	}
	return nil
}

func getAccountDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return home + "/.iwallet", nil
}

// LoadKeyPair ...
func LoadKeyPair(name string) (*account.KeyPair, error) {
	if name == "" {
		return nil, fmt.Errorf("you must provide account name")
	}
	dir, err := getAccountDir()
	if err != nil {
		return nil, err
	}
	privKeyFile := fmt.Sprintf("%s/%s_%s", dir, name, signAlgo)
	return sdk.LoadKeyPair(privKeyFile, signAlgo)
}

// InitAccount load account from file
func InitAccount() error {
	return LoadAndSetAccountForSDK(iwalletSDK)
}

// LoadAndSetAccountForSDK ...
func LoadAndSetAccountForSDK(s *sdk.IOSTDevSDK) error {
	keyPair, err := LoadKeyPair(accountName)
	if err != nil {
		return err
	}
	s.SetAccount(accountName, keyPair)
	return nil
}

// SaveAccount save account to file
func SaveAccount(name string, kp *account.KeyPair) error {
	dir, err := getAccountDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}
	fileName := dir + "/" + name
	if kp.Algorithm == crypto.Ed25519 {
		fileName += "_ed25519"
	}
	if kp.Algorithm == crypto.Secp256k1 {
		fileName += "_secp256k1"
	}

	pubfile, err := os.Create(fileName + ".pub")
	if err != nil {
		return fmt.Errorf("create file %v err %v", fileName+".pub", err)
	}
	defer pubfile.Close()

	_, err = pubfile.WriteString(common.Base58Encode(kp.Pubkey))
	if err != nil {
		return err
	}

	secFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
	if err != nil {
		return fmt.Errorf("create file %v err %v", fileName, err)
	}
	defer secFile.Close()

	_, err = secFile.WriteString(common.Base58Encode(kp.Seckey))
	if err != nil {
		return err
	}

	fmt.Println("Your account private key is saved at:", fileName)
	return nil
}

func actionsFromFlags(args []string) ([]*rpcpb.Action, error) {
	argc := len(args)
	if argc%3 != 0 {
		return nil, fmt.Errorf(`number of args should be a multiplier of 3`)
	}
	var actions = make([]*rpcpb.Action, 0)
	for i := 0; i < len(args); i += 3 {
		act := sdk.NewAction(args[i], args[i+1], args[i+2]) // Add some checks here.
		actions = append(actions, act)
	}
	return actions, nil
}

func checkFloat(cmd *cobra.Command, arg string, argName string) error {
	_, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		cmd.Usage()
		return fmt.Errorf(`invalid value "%v" for argument "%v": %v`, arg, argName, err)
	}
	return nil
}

func handleMultiSig(t *rpcpb.TransactionRequest, withSigns []string, signKeys []string) error {
	sigs := make([]*rpcpb.Signature, 0)
	if len(withSigns) != 0 && len(signKeys) != 0 {
		return fmt.Errorf("at least one of --sign_keys and --with_signs should be empty")
	}
	if len(signKeys) > 0 {
		for _, f := range signKeys {
			kp, err := sdk.LoadKeyPair(f, signAlgo)
			if err != nil {
				return fmt.Errorf("sign tx with priv key %v err %v", f, err)
			}
			sigs = append(sigs, sdk.GetSignatureOfTx(t, kp))
		}
	} else if len(withSigns) > 0 {
		for _, f := range withSigns {
			sig := &rpcpb.Signature{}
			err := sdk.LoadProtoStructFromJSONFile(f, sig)
			if err != nil {
				return fmt.Errorf("invalid signature file %v", f)
			}
			if !sdk.VerifySigForTx(t, sig) {
				return fmt.Errorf("sign verify error %v", f)
			}
			sigs = append(sigs, sig)
		}
	}
	t.Signatures = sigs
	return nil
}

// ParseAmountLimit ...
func ParseAmountLimit(limitStr string) ([]*rpcpb.AmountLimit, error) {
	result := make([]*rpcpb.AmountLimit, 0)
	if limitStr == "" {
		return result, nil
	}
	splits := strings.Split(limitStr, "|")
	for _, gram := range splits {
		limit := strings.Split(gram, ":")
		if len(limit) != 2 {
			return nil, fmt.Errorf("invalid amount limit %v", gram)
		}
		token := limit[0]
		if limit[1] != "unlimited" {
			amountLimit, err := strconv.ParseFloat(limit[1], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid amount limit %v %v", amountLimit, err)
			}
		}
		tokenLimit := &rpcpb.AmountLimit{}
		tokenLimit.Token = token
		tokenLimit.Value = limit[1]
		result = append(result, tokenLimit)
	}
	return result, nil
}
