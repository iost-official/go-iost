package iwallet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
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
	if accountName == "" {
		return errorWithHelp(cmd, "please provide the account name with flag --account/-a")
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

func parseTimeFromStr(s string) (int64, error) {
	var t time.Time
	if s == "" {
		return time.Now().UnixNano(), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, fmt.Errorf(`invalid time "%v", should in format "%v"`, s, time.Now().Format(time.RFC3339))
	}
	return t.UnixNano(), nil
}

func initTxFromActions(actions []*rpcpb.Action) (*rpcpb.TransactionRequest, error) {
	tx, err := iwalletSDK.CreateTxFromActions(actions)
	if err != nil {
		return nil, err
	}

	t, err := parseTimeFromStr(txTime)
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

func sendTxGetHash(tx *rpcpb.TransactionRequest) (string, error) {
	if err := InitAccount(); err != nil {
		return "", err
	}
	if !iwalletSDK.Connected() {
		if err := iwalletSDK.Connect(); err != nil {
			return "", err
		}
		defer iwalletSDK.CloseConn()
	}
	if err := handleMultiSig(tx, signatureFiles, signKeyFiles); err != nil {
		return "", err
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

func getAccountDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return home + "/.iwallet", nil
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

func loadAccountByName(name string, ensureDecrypt bool) (*AccountInfo, error) {
	accountDir, err := getAccountDir()
	if err != nil {
		return nil, err
	}
	fileName := accountDir + "/" + name + ".json"
	if _, err := os.Stat(fileName); err == nil {
		return loadAccountFromKeyStore(fileName, ensureDecrypt)
	}
	for _, algo := range ValidSignAlgos {
		fileName := accountDir + "/" + name + "_" + algo
		if _, err := os.Stat(fileName); err == nil {
			return loadAccountFromKeyPair(fileName)
		}
	}
	return nil, fmt.Errorf("account not exist")
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
	a, err := loadAccountByName(accountName, true)
	if err != nil {
		return err
	}
	kp, ok := a.Keypairs[signPerm]
	if !ok {
		return fmt.Errorf("invalid permission %v", signPerm)
	}
	keyPair, err := kp.toKeyPair()
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

func handleMultiSig(tx *rpcpb.TransactionRequest, signatureFiles []string, signKeyFiles []string) error {
	if len(signatureFiles) == 0 && len(signKeyFiles) == 0 {
		return nil
	}
	sigs := make([]*rpcpb.Signature, 0)
	if len(signatureFiles) != 0 && len(signKeyFiles) != 0 {
		return fmt.Errorf("can not set flags --sign_key_files and --signature_files simultaneously")
	}
	if len(signKeyFiles) > 0 {
		for _, f := range signKeyFiles {
			kp, err := sdk.LoadKeyPair(f, signAlgo)
			if err != nil {
				return fmt.Errorf("failed to sign tx with private key file %v: %v", f, err)
			}
			sigs = append(sigs, sdk.GetSignatureOfTx(tx, kp))
			fmt.Println("Signed transaction with private key file:", f)
		}
	} else if len(signatureFiles) > 0 {
		for _, f := range signatureFiles {
			sig := &rpcpb.Signature{}
			err := sdk.LoadProtoStructFromJSONFile(f, sig)
			if err != nil {
				return err
			}
			if !sdk.VerifySigForTx(tx, sig) {
				return fmt.Errorf("signature contained in %v is invalid", f)
			}
			sigs = append(sigs, sig)
			fmt.Println("Successfully added signature contained in:", f)
		}
	}
	tx.Signatures = sigs
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

func getAccountNameFromKeyPath(file string, suf string) (string, error) {
	f := file
	startIndex := strings.LastIndex(f, "/")
	//if startIndex == -1 {
	//	return "", fmt.Errorf("file name error, no '/' in %v", f)
	//}

	lastIndex := strings.LastIndex(f, suf)
	if lastIndex == -1 {
		return "", fmt.Errorf("file name error, no %v in %v", suf, f)
	}

	return f[startIndex+1 : lastIndex], nil
}

func getFilesAndDirs(dirPth string, suf string) (files []string, err error) { // nolint
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if !fi.IsDir() {
			ok := strings.HasSuffix(fi.Name(), suf)
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	return files, nil
}
