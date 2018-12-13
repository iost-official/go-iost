package gobang

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/iost-official/go-iost/iwallet"
	"github.com/iost-official/go-iost/test/performance/call"

	"encoding/json"

	"math/rand"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/rpc/pb"
	"google.golang.org/grpc"
)

var conns []*grpc.ClientConn
var rootKey = "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"
var rootID = "admin"
var rootAcc *account.KeyPair
var contractID string
var sdk = iwallet.SDK{}
var testID = "i" + strconv.FormatInt(time.Now().Unix(), 10)
var testAcc *account.KeyPair

type gobangHandle struct{}

func init() {
	gobang := new(gobangHandle)
	call.Register("gobang", gobang)
	rootAcc, _ = account.NewKeyPair(loadBytes(rootKey), crypto.Ed25519)
	rand.Seed(0)
}

// Init ...
func (t *gobangHandle) Init(add string, conNum int) error {
	initConn(add, conNum)
	return nil
}

// Publish ...
func (t *gobangHandle) Publish() error {
	var err error
	codePath := os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/test/performance/handles/gobang/gobang.js"
	abiPath := codePath + ".abi"
	sdk.SetAccount("admin", rootAcc)
	sdk.SetTxInfo(1000000, 100, 90, 0)
	sdk.SetCheckResult(true, 3, 10)
	testAcc, err = account.NewKeyPair(nil, crypto.Ed25519)
	if err != nil {
		return err
	}
	err = sdk.PledgeForGasAndRam(1500000, 0)
	if err != nil {
		return err
	}
	err = sdk.PledgeForGasAndRam(100, 10000000)
	if err != nil {
		return err
	}
	err = sdk.CreateNewAccount(testID, testAcc, 1000000, 10000, 100000)
	if err != nil {
		return err
	}
	sdk.SetAccount(testID, testAcc)
	_, txHash, err := sdk.PublishContract(codePath, abiPath, "", false, "")
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(5) * time.Second)
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

func (t *gobangHandle) wait(i int, h string, id string) bool {
	ilog.Info(id)
	g := rpcpb.GetContractStorageRequest{Id: contractID, Key: "games" + id, Field: "", ByLongestChain: true}
	for {
		ilog.Info("get storage, ", i)
		v, err := sdk.GetContractStorage(&g)
		if err != nil {
			ilog.Error(err)
		}
		var f interface{}
		err = json.Unmarshal([]byte(v.Data), &f)
		if err != nil {
			ilog.Error(err)
		}
		if f == nil {
			ilog.Info("f is nil, ", i)
			continue
		}
		m := f.(map[string]interface{})
		ilog.Info(m["hash"], " ", i)
		if m["hash"] == h {
			if m["winner"] != nil {
				return true
			}
			return false
		}
		time.Sleep(100 * time.Millisecond)
		ilog.Info(t.getGameID(h))
	}
}

func (t *gobangHandle) getGameID(h string) string {
	for {
		v, err := sdk.GetTxReceiptByTxHash(h)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			ilog.Info("nil")
			continue
		}
		ilog.Info(v)
		if len(v.Returns) == 0 {
			continue
		}
		var f interface{}
		err = json.Unmarshal([]byte(v.Returns[0]), &f)
		if err != nil {
			ilog.Error(err)
		}
		if f == nil {
			continue
		}
		m := f.([]interface{})
		return m[0].(string)
	}
}

func (t *gobangHandle) Transfer(i int) string {
	var board = make(map[string]bool)
	ilog.Info("transfer: ", i)
	act := tx.NewAction(contractID, "newGameWith", fmt.Sprintf(`["%v"]`, testID))
	h := t.transfer(i, act, rootAcc, rootID)
	gameID := t.getGameID(h)
	t.wait(i, h, gameID)
	round := 0
	for {
		acc := rootAcc
		id := rootID
		var x, y int
		for {
			x = rand.Intn(15)
			y = rand.Intn(15)
			if board[strconv.Itoa(x)+","+strconv.Itoa(y)] == false {
				board[strconv.Itoa(x)+","+strconv.Itoa(y)] = true
				break
			}
		}
		ilog.Info(i, " ", round, " ", x, " ", y)
		if round%2 == 1 {
			acc = testAcc
			id = testID
		}
		act = tx.NewAction(contractID, "move", fmt.Sprintf(`[%v, %v, %v, "%v"]`, gameID, x, y, h))
		h = t.transfer(i, act, acc, id)
		r := t.wait(i, h, gameID)
		if r {
			ilog.Info("end of game")
			break
		}
		round++
	}
	return ""
}

// Transfer ...
func (t *gobangHandle) transfer(i int, act *tx.Action, acc *account.KeyPair, id string) string {
	trx := tx.NewTx([]*tx.Action{act}, []string{}, 1000000000, 100, time.Now().Add(time.Second*time.Duration(10000)).UnixNano(), 0)
	stx, err := tx.SignTx(trx, id, []*account.KeyPair{acc})
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
