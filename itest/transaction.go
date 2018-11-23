package itest

import (
	"math"
	"strconv"
	"time"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/rpc/pb"
)

// Constant of Transaction
var (
	GasLimit   = int64(10000)           // about 2000~10000 gas per tx
	GasRatio   = int64(100)             // 1 mutiple gas
	Expiration = int64(math.MaxInt64)   // Max expired time is 90 seconds
	Delay      = int64(0 * time.Second) // No delay
	Signers    = make([]string, 0)      // No mutiple signers
)

// Transaction is the transaction object
type Transaction struct {
	*tx.Tx
}

// NewTransaction will return a new transaction by actions
func NewTransaction(actions []*tx.Action) *Transaction {
	t := tx.NewTx(
		actions,
		Signers,
		GasLimit,
		GasRatio,
		Expiration,
		Delay,
	)

	return &Transaction{t}
}

// NewTransactionFromPb returns a new transaction instance from protobuffer transaction struct.
func NewTransactionFromPb(t *rpcpb.Transaction) *Transaction {
	ret := &tx.Tx{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasRatio:   int64(t.GasRatio * 100),
		GasLimit:   int64(t.GasLimit * 100),
		Delay:      t.Delay,
		Signers:    t.Signers,
		Publisher:  t.Publisher,
	}
	for _, a := range t.Actions {
		ret.Actions = append(ret.Actions, &tx.Action{
			Contract:   a.Contract,
			ActionName: a.ActionName,
			Data:       a.Data,
		})
	}
	for _, a := range t.AmountLimit {
		ret.AmountLimit = append(ret.AmountLimit, &contract.Amount{
			Token: a.Token,
			Val:   strconv.FormatFloat(a.Value, 'f', -1, 64),
		})
	}
	return &Transaction{ret}
}

// ToTxRequest converts tx to rpcpb.TransactionRequest.
func (t *Transaction) ToTxRequest() *rpcpb.TransactionRequest {
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
