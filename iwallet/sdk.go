package iwallet

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/mitchellh/go-homedir"
	"google.golang.org/grpc"
)

// SDK ...
type SDK struct {
	server      string
	accountName string
	keyPair     *account.KeyPair
	signAlgo    string

	gasLimit    float64
	gasRatio    float64
	expiration  int64
	amountLimit string
	delaySecond int64

	checkResult         bool
	checkResultDelay    float32
	checkResultMaxRetry int32
	useLongestChain     bool

	verbose bool
}

var sdk = &SDK{}

// SetAccount ...
func (s *SDK) SetAccount(name string, kp *account.KeyPair) {
	s.accountName = name
	s.keyPair = kp
}

// SetTxInfo ...
func (s *SDK) SetTxInfo(gasLimit float64, gasRatio float64, expiration int64, delaySecond int64) {
	s.gasLimit = gasLimit
	s.gasRatio = gasRatio
	s.expiration = expiration
	s.delaySecond = delaySecond
}

// SetCheckResult ...
func (s *SDK) SetCheckResult(checkResult bool, checkResultDelay float32, checkResultMaxRetry int32) {
	s.checkResult = checkResult
	s.checkResultDelay = checkResultDelay
	s.checkResultMaxRetry = checkResultMaxRetry
}

// SetServer ...
func (s *SDK) SetServer(server string) {
	s.server = server
}

func (s *SDK) parseAmountLimit(limitStr string) ([]*rpcpb.AmountLimit, error) {
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

func (s *SDK) createTx(actions []*rpcpb.Action) (*rpcpb.TransactionRequest, error) {
	if s.amountLimit == "" {
		return nil, fmt.Errorf("cmdline flag --amountLimit must be set like `iost:300.00|ram:2000`. You can set to `*:unlimited` to disable any limit")
	}
	amountLimits, err := s.parseAmountLimit(s.amountLimit)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixNano()
	expiration := now + s.expiration*1e9

	ret := &rpcpb.TransactionRequest{
		Time:          time.Now().UnixNano(),
		Actions:       actions,
		Signers:       []string{},
		GasLimit:      s.gasLimit,
		GasRatio:      s.gasRatio,
		Expiration:    expiration,
		PublisherSigs: []*rpcpb.Signature{},
		Delay:         s.delaySecond * 1e9,
		AmountLimit:   amountLimits,
	}
	return ret, nil
}

func (s *SDK) signTx(t *rpcpb.TransactionRequest) (*rpcpb.TransactionRequest, error) {
	sig := &rpcpb.Signature{
		Algorithm: rpcpb.Signature_Algorithm(s.getSignAlgo()),
		Signature: s.getSignAlgo().Sign(common.Sha3(txToBytes(t)), s.keyPair.Seckey),
		PublicKey: s.getSignAlgo().GetPubkey(s.keyPair.Seckey),
	}
	t.PublisherSigs = []*rpcpb.Signature{sig}
	t.Publisher = s.accountName
	return t, nil
}

func (s *SDK) getSignAlgoName() string {
	return s.signAlgo
}

func (s *SDK) getSignAlgo() crypto.Algorithm {
	switch s.getSignAlgoName() {
	case "secp256k1":
		return crypto.Secp256k1
	case "ed25519":
		return crypto.Ed25519
	default:
		return crypto.Ed25519
	}
}

func (s *SDK) checkID(ID string) bool {
	if strings.HasPrefix(ID, "IOST") {
		return true
	}
	return false
}

// GetContractStorage ...
func (s *SDK) GetContractStorage(r *rpcpb.GetContractStorageRequest) (*rpcpb.GetContractStorageResponse, error) {
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	value, err := client.GetContractStorage(context.Background(), r)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *SDK) getNodeInfo() (*rpcpb.NodeInfoResponse, error) {
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	value, err := client.GetNodeInfo(context.Background(), &rpcpb.EmptyRequest{})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *SDK) getChainInfo() (*rpcpb.ChainInfoResponse, error) {
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	value, err := client.GetChainInfo(context.Background(), &rpcpb.EmptyRequest{})
	if err != nil {
		return nil, err
	}
	return value, nil
}

// getAccountInfo return account info
func (s *SDK) getAccountInfo(id string) (*rpcpb.Account, error) {
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	req := &rpcpb.GetAccountRequest{Name: id, ByLongestChain: s.useLongestChain}
	value, err := client.GetAccount(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return value, nil
}
func (s *SDK) getGetBlockByNum(num int64, complete bool) (*rpcpb.BlockResponse, error) {
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	return client.GetBlockByNumber(context.Background(), &rpcpb.GetBlockByNumberRequest{Number: num, Complete: complete})
}

func (s *SDK) getGetBlockByHash(hash string, complete bool) (*rpcpb.BlockResponse, error) {
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	return client.GetBlockByHash(context.Background(), &rpcpb.GetBlockByHashRequest{Hash: hash, Complete: complete})
}

func (s *SDK) getTxByHash(hash string) (*rpcpb.TransactionResponse, error) {
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	return client.GetTxByHash(context.Background(), &rpcpb.TxHashRequest{Hash: hash})
}

// GetTxReceiptByTxHash ...
func (s *SDK) GetTxReceiptByTxHash(txHashStr string) (*rpcpb.TxReceipt, error) {
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	return client.GetTxReceiptByTxHash(context.Background(), &rpcpb.TxHashRequest{Hash: txHashStr})
}

func (s *SDK) sendTx(stx *rpcpb.TransactionRequest) (string, error) {
	fmt.Println("sending tx")
	if sdk.verbose {
		fmt.Println(stx.String())
	}
	conn, err := grpc.Dial(s.server, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()
	client := rpcpb.NewApiServiceClient(conn)
	resp, err := client.SendTransaction(context.Background(), stx)
	if err != nil {
		return "", err
	}
	return resp.Hash, nil
}

func (s *SDK) checkTransaction(txHash string) bool {
	// It may be better to to create a grpc client and reuse it. TODO later
	for i := int32(0); i < s.checkResultMaxRetry; i++ {
		time.Sleep(time.Duration(s.checkResultDelay*1000) * time.Millisecond)
		txReceipt, err := s.GetTxReceiptByTxHash(txHash)
		if err != nil {
			fmt.Println("result not ready, please wait. Details: ", err)
			continue
		}
		if txReceipt == nil {
			fmt.Println("result not ready, please wait.")
			continue
		}
		if txReceipt.StatusCode != rpcpb.TxReceipt_SUCCESS {
			fmt.Println("exec tx failed: ", txReceipt.Message)
			fmt.Println("full error information: ", marshalTextString(txReceipt))
		} else {
			fmt.Println("exec tx done")
			if s.verbose {
				fmt.Println(marshalTextString(txReceipt))
			}
			return true
		}
		break
	}
	return false
}

func (s *SDK) getAccountDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return home + "/.iwallet", nil
}

func (s *SDK) loadAccount() error {
	dir, err := s.getAccountDir()
	if err != nil {
		return err
	}
	if s.accountName == "" {
		return fmt.Errorf("you must provide account name")
	}
	kpPath := fmt.Sprintf("%s/%s_%s", dir, s.accountName, s.getSignAlgoName())
	fsk, err := loadKey(kpPath)
	if err != nil {
		return fmt.Errorf("read file failed: %v", err)
	}
	s.keyPair, err = account.NewKeyPair(loadBytes(string(fsk)), s.getSignAlgo())
	if err != nil {
		return err
	}
	return nil
}

func (s *SDK) saveAccount(name string, kp *account.KeyPair) error {
	dir, err := s.getAccountDir()
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
		return err
	}
	defer pubfile.Close()

	_, err = pubfile.WriteString(saveBytes(kp.Pubkey))
	if err != nil {
		return err
	}

	secFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
	if err != nil {
		return err
	}
	defer secFile.Close()

	_, err = secFile.WriteString(saveBytes(kp.Seckey))
	if err != nil {
		return err
	}

	idFileName := fileName + ".id"
	idFile, err := os.Create(idFileName)
	if err != nil {
		return err
	}
	defer idFile.Close()
	id := account.GetIDByPubkey(kp.Pubkey)
	_, err = idFile.WriteString(id)
	if err != nil {
		return err
	}

	//fmt.Println("your account id is saved at:")
	//fmt.Println(idFileName)
	fmt.Println("your account private key is saved at:")
	fmt.Println(fileName)
	return nil
}

// PledgeForGasAndRAM ...
func (s *SDK) PledgeForGasAndRAM(gasPledged int64, ram int64) error {
	var acts []*rpcpb.Action
	acts = append(acts, NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, s.accountName, s.accountName, gasPledged)))
	if ram > 0 {
		acts = append(acts, NewAction("ram.iost", "buy", fmt.Sprintf(`["%v", "%v", %v]`, s.accountName, s.accountName, ram)))
	}
	trx, err := s.createTx(acts)
	if err != nil {
		return err
	}
	stx, err := s.signTx(trx)
	if err != nil {
		return err
	}
	var txHash string
	txHash, err = s.sendTx(stx)
	if err != nil {
		return err
	}
	fmt.Printf("send tx done\n")
	if s.checkResult {
		s.checkTransaction(txHash)
	}
	fmt.Printf("\nbalance of %v\n", s.accountName)
	info, err := s.getAccountInfo(s.accountName)
	if err != nil {
		return err
	}
	fmt.Println(marshalTextString(info))
	return nil
}

// CreateNewAccount ...
func (s *SDK) CreateNewAccount(newID string, ownerKey string, activeKey string, initialGasPledge int64, initialRAM int64, initialCoins int64) error {
	var acts []*rpcpb.Action
	acts = append(acts, NewAction("auth.iost", "SignUp", fmt.Sprintf(`["%v", "%v", "%v"]`, newID, ownerKey, activeKey)))
	if initialRAM > 0 {
		acts = append(acts, NewAction("ram.iost", "buy", fmt.Sprintf(`["%v", "%v", %v]`, s.accountName, newID, initialRAM)))
	}
	if initialGasPledge > 0 {
		acts = append(acts, NewAction("gas.iost", "pledge", fmt.Sprintf(`["%v", "%v", "%v"]`, s.accountName, newID, initialGasPledge)))
	}
	if initialCoins > 0 {
		acts = append(acts, NewAction("token.iost", "transfer", fmt.Sprintf(`["iost", "%v", "%v", "%v", ""]`, s.accountName, newID, initialCoins)))
	}
	trx, err := s.createTx(acts)
	if err != nil {
		return err
	}
	stx, err := s.signTx(trx)
	if err != nil {
		return err
	}
	var txHash string
	txHash, err = s.sendTx(stx)
	if err != nil {
		return err
	}
	fmt.Println("send tx done")
	fmt.Println("the create user transaction hash is:", txHash)
	if s.checkResult {
		if !s.checkTransaction(txHash) {
			return fmt.Errorf("create new account %v transaction failed", newID)
		}
	}
	fmt.Printf("balance of %v\n", newID)
	info, err := s.getAccountInfo(newID)
	if err != nil {
		return err
	}
	fmt.Println(marshalTextString(info))
	return nil
}

// PublishContract converts contract js code to transaction. If 'send', also send it to chain.
func (s *SDK) PublishContract(codePath string, abiPath string, conID string, update bool, updateID string) (stx *rpcpb.TransactionRequest, txHash string, err error) {
	fd, err := readFile(codePath)
	if err != nil {
		fmt.Println("Read source code file failed: ", err.Error())
		return nil, "", err
	}
	code := string(fd)

	fd, err = readFile(abiPath)
	if err != nil {
		fmt.Println("Read abi file failed: ", err.Error())
		return nil, "", err
	}
	abi := string(fd)
	var info *contract.Info
	err = json.Unmarshal([]byte(abi), &info)
	if err != nil {
		return nil, "", err
	}
	c := &contract.Contract{
		ID:   conID,
		Code: code,
		Info: info,
	}
	methodName := "SetCode"
	if update {
		methodName = "UpdateCode"
	}
	marshalMethod := "json"
	var contractStr string
	if marshalMethod == "json" {
		buf, err := json.Marshal(c)
		if err != nil {
			return nil, "", err
		}
		contractStr = string(buf)
	} else {
		buf, err := proto.Marshal(c)
		if err != nil {
			return nil, "", err
		}
		contractStr = base64.StdEncoding.EncodeToString(buf)
	}
	arr := []string{contractStr}
	if update {
		arr = append(arr, updateID)
	}
	data, err := json.Marshal(arr)
	if err != nil {
		return nil, "", err
	}
	action := NewAction("system.iost", methodName, string(data))
	trx, err := s.createTx([]*rpcpb.Action{action})
	if err != nil {
		return nil, "", err
	}
	stx, err = s.signTx(trx)
	if err != nil {
		return nil, "", fmt.Errorf("sign tx error %v", err)
	}
	var hash string
	hash, err = s.sendTx(stx)
	if err != nil {
		return nil, "", err
	}
	fmt.Println("Sending tx to rpc server finished. The transaction hash is:", hash)
	return trx, hash, nil
}

func actionToBytes(a *rpcpb.Action) []byte {
	sn := common.NewSimpleNotation()
	sn.WriteString(a.Contract, true)
	sn.WriteString(a.ActionName, true)
	sn.WriteString(a.Data, true)
	return sn.Bytes()
}

func amountToBytes(a *rpcpb.AmountLimit) []byte {
	sn := common.NewSimpleNotation()
	sn.WriteString(a.Token, true)
	sn.WriteString(a.Value, true)
	return sn.Bytes()
}

func signatureToBytes(s *rpcpb.Signature) []byte {
	sn := common.NewSimpleNotation()
	sn.WriteByte(byte(s.Algorithm), true)
	sn.WriteBytes(s.Signature, true)
	sn.WriteBytes(s.PublicKey, true)
	return sn.Bytes()
}

func txToBytes(t *rpcpb.TransactionRequest) []byte {
	sn := common.NewSimpleNotation()
	sn.WriteInt64(t.Time, true)
	sn.WriteInt64(t.Expiration, true)
	sn.WriteInt64(int64(t.GasRatio*100), true)
	sn.WriteInt64(int64(t.GasLimit*100), true)
	sn.WriteInt64(t.Delay, true)
	sn.WriteStringSlice(t.Signers, true)

	actionBytes := make([][]byte, 0, len(t.Actions))
	for _, a := range t.Actions {
		actionBytes = append(actionBytes, actionToBytes(a))
	}
	sn.WriteBytesSlice(actionBytes, false)

	amountBytes := make([][]byte, 0, len(t.AmountLimit))
	for _, a := range t.AmountLimit {
		amountBytes = append(amountBytes, amountToBytes(a))
	}
	sn.WriteBytesSlice(amountBytes, false)

	signBytes := make([][]byte, 0, len(t.Signatures))
	for _, sig := range t.Signatures {
		signBytes = append(signBytes, signatureToBytes(sig))
	}
	sn.WriteBytesSlice(signBytes, false)

	return sn.Bytes()
}

// NewAction ...
func NewAction(contract string, name string, data string) *rpcpb.Action {
	return &rpcpb.Action{
		Contract:   contract,
		ActionName: name,
		Data:       data,
	}
}
