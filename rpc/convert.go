package rpc

import (
	"encoding/json"
	"strconv"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	contract "github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/rpc/pb"
)

func gasConvert(a int64) float64 {
	return float64(a) / 100
}

func fixToFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func toPbAction(a *tx.Action) *rpcpb.Action {
	return &rpcpb.Action{
		Contract:   a.Contract,
		ActionName: a.ActionName,
		Data:       a.Data,
	}
}

func toPbTxReceipt(tr *tx.TxReceipt) *rpcpb.TxReceipt {
	if tr == nil {
		return nil
	}
	ret := &rpcpb.TxReceipt{
		TxHash:     common.Base58Encode(tr.TxHash),
		GasUsage:   gasConvert(tr.GasUsage),
		RamUsage:   tr.RAMUsage,
		StatusCode: rpcpb.TxReceipt_StatusCode(tr.Status.Code),
		Message:    tr.Status.Message,
		Returns:    tr.Returns,
	}
	for _, r := range tr.Receipts {
		ret.Receipts = append(ret.Receipts, &rpcpb.TxReceipt_Receipt{
			FuncName: r.FuncName,
			Content:  r.Content,
		})
	}
	return ret
}

func toPbAmountLimit(a *contract.Amount) *rpcpb.AmountLimit {
	return &rpcpb.AmountLimit{
		Token: a.Token,
		Value: a.Val,
	}
}

func toPbTx(t *tx.Tx, tr *tx.TxReceipt) *rpcpb.Transaction {
	ret := &rpcpb.Transaction{
		Hash:       common.Base58Encode(t.Hash()),
		Time:       t.Time,
		Expiration: t.Expiration,
		GasPrice:   gasConvert(t.GasPrice),
		GasLimit:   gasConvert(t.GasLimit),
		Delay:      t.Delay,
		Signers:    t.Signers,
		Publisher:  t.Publisher,
		ReferredTx: common.Base58Encode(t.ReferredTx),
		TxReceipt:  toPbTxReceipt(tr),
	}
	for _, a := range t.Actions {
		ret.Actions = append(ret.Actions, toPbAction(a))
	}
	for _, a := range t.AmountLimit {
		ret.AmountLimit = append(ret.AmountLimit, toPbAmountLimit(a))
	}
	return ret
}

func toPbBlock(blk *block.Block, complete bool) *rpcpb.Block {
	ret := &rpcpb.Block{
		Hash:                common.Base58Encode(blk.HeadHash()),
		Version:             blk.Head.Version,
		ParentHash:          common.Base58Encode(blk.Head.ParentHash),
		TxMerkleHash:        common.Base58Encode(blk.Head.TxMerkleHash),
		TxReceiptMerkleHash: common.Base58Encode(blk.Head.TxReceiptMerkleHash),
		Number:              blk.Head.Number,
		Witness:             blk.Head.Witness,
		Time:                blk.Head.Time,
		GasUsage:            gasConvert(blk.Head.GasUsage),
		TxCount:             int64(len(blk.Txs)),
	}
	json.Unmarshal(blk.Head.Info, ret.Info)
	if complete {
		for i, t := range blk.Txs {
			ret.Transactions = append(ret.Transactions, toPbTx(t, blk.Receipts[i]))
		}
	}
	return ret
}
