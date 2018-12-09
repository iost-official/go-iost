package gobang

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/iost-official/go-iost/iwallet"
	"github.com/iost-official/go-iost/test/performance/call"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
	"google.golang.org/grpc"
)

var conns []*grpc.ClientConn
var rootKey = "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"
var contractID string
var sdk = iwallet.SDK{}

var testID = "i" + strconv.FormatInt(time.Now().Unix(), 10)

type transferHandle struct{}

func init() {
	transfer := new(transferHandle)
	call.Register("transfer", transfer)
}

// Init ...
func (t *transferHandle) Init(add string, conNum int) error {
	initConn(add, conNum)
	return nil
}

// Publish ...
func (t *transferHandle) Publish() error {
	acc, _ := account.NewKeyPair(loadBytes(rootKey), crypto.Ed25519)
	codePath := os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/test/performance/handles/transfer/transfer.js"
	abiPath := codePath + ".abi"
	sdk.SetAccount("admin", acc)
	sdk.SetTxInfo(5000000, 100, 90, 0)
	sdk.SetCheckResult(true, 3, 10)
	testKp, err := account.NewKeyPair(nil, crypto.Ed25519)
	if err != nil {
		return err
	}
	err = sdk.CreateNewAccount(testID, testKp, 1000000, 10000, 100000)
	if err != nil {
		return err
	}
	err = sdk.PledgeForGas(1500000)
	if err != nil {
		return err
	}
	sdk.SetAccount(testID, testKp)
	_, txHash, err := sdk.PublishContract(codePath, abiPath, "", false, "")
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(30) * time.Second)
	client := rpcpb.NewApiServiceClient(conns[0])
	resp, err := client.GetTxReceiptByTxHash(context.Background(), &rpcpb.TxHashRequest{Hash: txHash})
	if err != nil {
		return err
	}
	if tx.StatusCode(resp.StatusCode) != tx.Success {
		return fmt.Errorf("publish contract fail " + (resp.String()))
	}

	contractID = "Contract" + txHash
	return nil
}

// Transfer ...
func (t *transferHandle) Transfer(i int) string {
	action := tx.NewAction(contractID, "transfer", fmt.Sprintf(`["admin","%v",1]`, testID))
	acc, _ := account.NewKeyPair(loadBytes(rootKey), crypto.Ed25519)
	trx := tx.NewTx([]*tx.Action{action}, []string{}, 5000000, 100, time.Now().Add(time.Second*time.Duration(10000)).UnixNano(), 0)
	stx, err := tx.SignTx(trx, "admin", []*account.KeyPair{acc})

	if err != nil {
		return fmt.Sprintf("signtx:%v err:%v", stx, err)
	}
	var txHash []byte
	txHash, err = sendTx(stx, i)
	if err != nil {
		return fmt.Sprintf("sendtx:%v  err:%v", txHash, err)
	}
	return string(txHash)
}

func initConn(add string, num int) {
	sdk.SetServer(add)
	conns = make([]*grpc.ClientConn, num)
	allServers := []string{add}
	for i := 0; i < num; i++ {
		conn, err := grpc.Dial(allServers[i%len(allServers)], grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		conns[i] = conn
	}
}

func toTxRequest(t *tx.Tx) *rpcpb.TransactionRequest {
	ret := &rpcpb.TransactionRequest{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasRatio:   float64(t.GasRatio) / 100,
		GasLimit:   float64(t.GasLimit) / 100,
		Delay:      t.Delay,
		Signers:    t.Signers,
		Publisher:  t.Publisher,
	}
	for _, a := range t.Actions {
		ret.Actions = append(ret.Actions, &rpcpb.Action{
			Contract:   a.Contract,
			ActionName: a.ActionName,
			Data:       a.Data,
		})
	}
	for _, a := range t.AmountLimit {
		fixed, err := common.UnmarshalFixed(a.Val)
		if err != nil {
			continue
		}
		ret.AmountLimit = append(ret.AmountLimit, &rpcpb.AmountLimit{
			Token: a.Token,
			Value: fixed.ToFloat(),
		})
	}
	for _, s := range t.Signs {
		ret.Signatures = append(ret.Signatures, &rpcpb.Signature{
			Algorithm: rpcpb.Signature_Algorithm(s.Algorithm),
			PublicKey: s.Pubkey,
			Signature: s.Sig,
		})
	}
	for _, s := range t.PublishSigns {
		ret.PublisherSigs = append(ret.PublisherSigs, &rpcpb.Signature{
			Algorithm: rpcpb.Signature_Algorithm(s.Algorithm),
			PublicKey: s.Pubkey,
			Signature: s.Sig,
		})
	}
	return ret
}

func sendTx(stx *tx.Tx, i int) ([]byte, error) {
	client := rpcpb.NewApiServiceClient(conns[i])
	resp, err := client.SendTransaction(context.Background(), toTxRequest(stx))
	if err != nil {
		return nil, err
	}
	return []byte(resp.Hash), nil
}

func loadBytes(s string) []byte {
	if s[len(s)-1] == 10 {
		s = s[:len(s)-1]
	}
	buf := common.Base58Decode(s)
	return buf
}
