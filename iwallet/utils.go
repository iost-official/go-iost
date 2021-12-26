package iwallet

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/spf13/cobra"

	rpcpb "github.com/iost-official/go-iost/v3/rpc/pb"
	"github.com/iost-official/go-iost/v3/sdk"
)

////////////////////// check input args ////////////////////

func errorWithHelp(cmd *cobra.Command, format string, a ...any) error {
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

///////////////////////// construct tx /////////////////////////

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

func createActionFromMethod(contract, method string, methodArgs ...any) (*rpcpb.Action, error) {
	methodArgsBytes, err := json.Marshal(methodArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal method args %v: %v", methodArgs, err)
	}
	return sdk.NewAction(contract, method, string(methodArgsBytes)), nil
}

func createTxFromActions(actions []*rpcpb.Action) (*rpcpb.TransactionRequest, error) {
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

func formatContractArgs(data string) (string, error) {
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
			accInfo, err := loadAccountFrom(f, true)
			if err != nil {
				return fmt.Errorf("failed to load account from file %v: %v", f, err)
			}
			kp, err := accInfo.GetKeyPair("avtive")
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

func prepareTx(tx *rpcpb.TransactionRequest) error {
	if iwalletSDK.CurrentAccount() == "" {
		if err := initAccountForSDK(iwalletSDK); err != nil {
			return err
		}
	}
	if err := handleMultiSig(tx, signatureFiles, signKeyFiles, asPublisherSign); err != nil {
		return err
	}
	if err := checkTxTime(tx); err != nil {
		return err
	}
	return nil
}

func sendTx(tx *rpcpb.TransactionRequest) (string, error) {
	err := prepareTx(tx)
	if err != nil {
		return "", err
	}
	return iwalletSDK.SendTx(tx)
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

func processTx(tx *rpcpb.TransactionRequest) (string, error) {
	if outputTxFile != "" {
		return "", saveTx(tx)
	}
	err := prepareTx(tx)
	if err != nil {
		return "", err
	}
	if tryTx {
		r, err := iwalletSDK.TryTx(tx)
		if err != nil {
			return "", err
		}
		fmt.Println(sdk.MarshalTextString(r))
		return "", nil
	}
	return iwalletSDK.SendTx(tx)
}

func processActions(actions []*rpcpb.Action) (string, error) {
	tx, err := createTxFromActions(actions)
	if err != nil {
		return "", err
	}
	return processTx(tx)
}

func processMethod(contract, method string, methodArgs ...any) error {
	action, err := createActionFromMethod(contract, method, methodArgs...)
	if err != nil {
		return err
	}
	_, err = processActions([]*rpcpb.Action{action})
	return err
}
