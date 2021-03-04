package iwallet

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"

	simplejson "github.com/bitly/go-simplejson"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	rpcpb "github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
)

func errorWithHelp(cmd *cobra.Command, format string, a ...interface{}) error {
	cmd.Help()
	fmt.Println()
	return fmt.Errorf(format, a...)
}

func checkArgsNumber(cmd *cobra.Command, args []string, argNames ...string) error {
	if len(args) < len(argNames) {
		return errorWithHelp(cmd, "missing positional argument: %v", argNames[len(args):])
	}
	return nil
}

func checkAccount(cmd *cobra.Command) error {
	if accountName == "" && keyFile == "" {
		return errorWithHelp(cmd, "please provide the account info with flag --account/-a or --key_file/-k")
	}
	return nil
}

func checkFloat(cmd *cobra.Command, arg string, argName string) error {
	_, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return errorWithHelp(cmd, `invalid value "%v" for argument "%v": %v`, arg, argName, err)
	}
	return nil
}

func checkSigners(signers []string) error {
	for _, s := range signers {
		if !(len(strings.Split(s, "@")) == 2) {
			return fmt.Errorf("signer %v should contain '@'", s)
		}
	}
	return nil
}

func parseTxTime() (int64, error) {
	if txTime != "" && txTimeDelay != 0 {
		return 0, fmt.Errorf("can not set flags --tx_time and --tx_time_delay simultaneously")
	}
	if txTime != "" {
		t, err := time.Parse(time.RFC3339, txTime)
		if err != nil {
			return 0, fmt.Errorf(`invalid time "%v", should in format "%v"`, txTime, time.Now().Format(time.RFC3339))
		}
		return t.UnixNano(), nil
	}
	t := time.Now()
	t = t.Add(time.Second * time.Duration(txTimeDelay))
	return t.UnixNano(), nil
}

func initTxFromActions(actions []*rpcpb.Action) (*rpcpb.TransactionRequest, error) {
	tx, err := iwalletSDK.CreateTxFromActions(actions)
	if err != nil {
		return nil, err
	}

	t, err := parseTxTime()
	if err != nil {
		return nil, err
	}
	tx.Time = t
	tx.Expiration = tx.Time + expiration*1e9

	if err := checkSigners(signers); err != nil {
		return nil, err
	}
	tx.Signers = signers

	return tx, nil
}

func initTxFromMethod(contract, method string, methodArgs ...interface{}) (*rpcpb.TransactionRequest, error) {
	methodArgsBytes, err := json.Marshal(methodArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal method args %v: %v", methodArgs, err)
	}
	action := sdk.NewAction(contract, method, string(methodArgsBytes))
	actions := []*rpcpb.Action{action}
	return initTxFromActions(actions)
}

func checkTxTime(tx *rpcpb.TransactionRequest) error {
	timepoint := time.Unix(0, tx.Time)
	delta := time.Until(timepoint)
	if math.Abs(delta.Seconds()) > 1 {
		fmt.Println("The transaction time is:", timepoint.Format(time.RFC3339))
	}
	if delta.Seconds() < 1 {
		return nil
	}
	seconds := int(delta.Seconds())
	if seconds%10 > 0 {
		fmt.Printf("Waiting %v seconds to reach the transaction time...\n", seconds)
		time.Sleep(time.Duration(seconds%10) * time.Second)
	}
	for i := seconds % 10; i < seconds; i += 10 {
		fmt.Printf("Waiting %v seconds to reach the transaction time...\n", seconds-i)
		time.Sleep(10 * time.Second)
	}
	return nil
}

func prepareTx(tx *rpcpb.TransactionRequest) error {
	if err := LoadAndSetAccountForSDK(iwalletSDK); err != nil {
		return err
	}
	if err := handleMultiSig(tx, signatureFiles, signKeyFiles, asPublisherSign); err != nil {
		return err
	}
	if err := checkTxTime(tx); err != nil {
		return err
	}
	return nil
}

func sendTxGetHash(tx *rpcpb.TransactionRequest) (string, error) {
	err := prepareTx(tx)
	if err != nil {
		return "", err
	}
	if !iwalletSDK.Connected() {
		if err := iwalletSDK.Connect(); err != nil {
			return "", err
		}
		defer iwalletSDK.CloseConn()
	}
	return iwalletSDK.SendTx(tx)
}

func sendTx(tx *rpcpb.TransactionRequest) error {
	_, err := sendTxGetHash(tx)
	return err
}

func saveTx(tx *rpcpb.TransactionRequest) error {
	err := sdk.SaveProtoStructToJSONFile(tx, outputTxFile)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Println("Transaction:")
		fmt.Println(sdk.MarshalTextString(tx))
	}
	fmt.Println("Successfully saved transaction request as json file:", outputTxFile)
	return nil
}

func saveOrSendTx(tx *rpcpb.TransactionRequest) error {
	if outputTxFile != "" {
		return saveTx(tx)
	}
	return sendTx(tx)
}

func saveOrSendAction(contract, method string, methodArgs ...interface{}) error {
	tx, err := initTxFromMethod(contract, method, methodArgs...)
	if err != nil {
		return err
	}
	return saveOrSendTx(tx)
}

// GetSignAlgoByName ...
func GetSignAlgoByName(name string) crypto.Algorithm {
	switch name {
	case "secp256k1":
		return crypto.Secp256k1
	case "ed25519":
		return crypto.Ed25519
	default:
		return crypto.Ed25519
	}
}

func loadAccount() (*AccountInfo, error) {
	if keyFile != "" {
		if accountDir != "" {
			ilog.Warn("--key_file is set, so --account_dir will be ignored")
		}
		acc, err := LoadAccountFromKeyStore(keyFile, true)
		if err != nil {
			return nil, err
		}
		if accountName != "" && acc.Name != accountName {
			return nil, fmt.Errorf("inconsistent account: %s from cmd args VS %s from key file", accountName, acc.Name)
		}
		if accountName == "" {
			accountName = acc.Name
		}
		return acc, nil
	}
	return loadAccountByName(accountName, true)
}

func getAccountDir() (string, error) {
	if accountDir != "" {
		return path.Join(accountDir, ".iwallet"), nil
	}
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return path.Join(home, ".iwallet"), nil
}

func loadAccountByName(name string, ensureDecrypt bool) (*AccountInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("account name should be provived by --account_name")
	}
	accountDir, err := getAccountDir()
	if err != nil {
		return nil, err
	}
	fileName := accountDir + "/" + name + ".json"
	_, err = os.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("account is not imported at %s: %v. use 'iwallet account import %s <private-key>' to import it", fileName, err, name)
	}
	return LoadAccountFromKeyStore(fileName, ensureDecrypt)
}

// SetAccountForSDK ...
func SetAccountForSDK(s *sdk.IOSTDevSDK, a *AccountInfo, signPerm string) error {
	kp, ok := a.Keypairs[signPerm]
	if !ok {
		return fmt.Errorf("invalid permission %v", signPerm)
	}
	keyPair, err := kp.toKeyPair()
	if err != nil {
		return err
	}
	s.SetAccount(a.Name, keyPair)
	s.SetSignAlgo(kp.KeyType)
	return nil
}

// LoadAndSetAccountForSDK load account from file
func LoadAndSetAccountForSDK(s *sdk.IOSTDevSDK) error {
	a, err := loadAccount()
	if err != nil {
		return err
	}
	return SetAccountForSDK(s, a, signPerm)
}

func argsFormatter(data string) (string, error) {
	if !strings.ContainsAny(data, "[]{}'\",") {
		// we treat xxx as '["xxx"]'
		return "[\"" + data + "\"]", nil
	}
	js, err := simplejson.NewJson([]byte(data))
	if err != nil {
		return "", fmt.Errorf("invalid args, should be json array: %v, %v", data, err)
	}
	_, err = js.Array()
	if err != nil {
		return "", fmt.Errorf("invalid args, should be json array: %v, %v", data, err)
	}
	b, err := js.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func actionsFromFlags(args []string) ([]*rpcpb.Action, error) {
	argc := len(args)
	if argc%3 != 0 {
		return nil, fmt.Errorf(`number of args should be a multiplier of 3`)
	}
	var actions = make([]*rpcpb.Action, 0)
	for i := 0; i < len(args); i += 3 {
		v, err := argsFormatter(args[i+2])
		if err != nil {
			return nil, err
		}
		act := sdk.NewAction(args[i], args[i+1], v)
		actions = append(actions, act)
	}
	return actions, nil
}

func handleMultiSig(tx *rpcpb.TransactionRequest, signatureFiles []string, signKeyFiles []string, asPublisherSign bool) error {
	if len(signatureFiles) == 0 && len(signKeyFiles) == 0 {
		return nil
	}
	sigs := make([]*rpcpb.Signature, 0)
	if len(signatureFiles) != 0 && len(signKeyFiles) != 0 {
		return fmt.Errorf("can not set flags --sign_key_files and --signature_files simultaneously")
	}
	if len(signKeyFiles) > 0 {
		for _, f := range signKeyFiles {
			accInfo, err := LoadAccountFromKeyStore(f, true)
			if err != nil {
				return fmt.Errorf("failed to load account from file %v: %v", f, err)
			}
			kp, err := accInfo.Keypairs["active"].toKeyPair()
			if err != nil {
				return fmt.Errorf("failed to get key pair from file %v: %v", f, err)
			}
			sigs = append(sigs, sdk.GetSignatureOfTx(tx, kp, asPublisherSign))
			fmt.Println("Signed transaction with private key file:", f)
		}
	} else if len(signatureFiles) > 0 {
		for _, f := range signatureFiles {
			sig := &rpcpb.Signature{}
			err := sdk.LoadProtoStructFromJSONFile(f, sig)
			if err != nil {
				return err
			}
			sigs = append(sigs, sig)
			fmt.Println("Successfully added signature contained in:", f)
		}
	}
	if asPublisherSign {
		tx.PublisherSigs = sigs
	} else {
		tx.Signatures = sigs
	}
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

// ValidSignAlgos ...
var ValidSignAlgos = []string{"ed25519", "secp256k1"}
