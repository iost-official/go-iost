package itest

import (
	"math"
	"time"

	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/rpc/pb"
)

// Constant of Transaction
var (
	GasLimit    = int64(100000000)       // about 30000~100000 gas per tx
	GasRatio    = int64(100)             // 1 mutiple gas
	Expiration  = int64(math.MaxInt64)   // Max expired time is 90 seconds
	Delay       = int64(0 * time.Second) // No delay
	Signers     = make([]string, 0)      // No mutiple signers
	AmountLimit = []*contract.Amount{{Token: "iost", Val: "unlimited"}}
	ChainID     uint32
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
		ChainID,
	)
	t.AmountLimit = AmountLimit

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
			Val:   a.Value,
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
		ret.AmountLimit = append(ret.AmountLimit, &rpcpb.AmountLimit{
			Token: a.Token,
			Value: a.Val,
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
