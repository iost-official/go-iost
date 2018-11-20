package rpc

import (
	"encoding/json"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc/pb"
)

func gasConvert(a int64) float64 {
	return float64(a) / 100
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
		GasUsage:            gasConvert(blk.CalculateGasUsage()),
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

func toPbItem(item *account.Item) *rpcpb.Account_Item {
	return &rpcpb.Account_Item{
		Id:         item.ID,
		Permission: item.Permission,
		IsKeyPair:  item.IsKeyPair,
		Weight:     int64(item.Weight),
	}
}

func toPbGroup(group *account.Group) *rpcpb.Account_Group {
	ret := &rpcpb.Account_Group{Name: group.Name}
	for _, u := range group.Users {
		ret.Items = append(ret.Items, toPbItem(u))
	}
	return ret
}

func toPbPermission(permission *account.Permission) *rpcpb.Account_Permission {
	ret := &rpcpb.Account_Permission{
		Name:      permission.Name,
		Groups:    permission.Groups,
		Threshold: int64(permission.Threshold),
	}
	for _, u := range permission.Users {
		ret.Items = append(ret.Items, toPbItem(u))
	}
	return ret
}

func toPbAccount(acc *account.Account) *rpcpb.Account {
	ret := &rpcpb.Account{
		Name:        acc.ID,
		Permissions: make(map[string]*rpcpb.Account_Permission),
		Groups:      make(map[string]*rpcpb.Account_Group),
	}
	for k, p := range acc.Permissions {
		ret.Permissions[k] = toPbPermission(p)
	}
	for k, g := range acc.Groups {
		ret.Groups[k] = toPbGroup(g)
	}
	return ret
}

func toPbContract(c *contract.Contract) *rpcpb.Contract {
	ret := &rpcpb.Contract{
		Id:       c.ID,
		Code:     c.Code,
		Language: c.Info.Lang,
		Version:  c.Info.Version,
	}
	for _, abi := range c.Info.Abi {
		pbABI := &rpcpb.Contract_ABI{
			Name: abi.Name,
			Args: abi.Args,
		}
		for _, al := range abi.AmountLimit {
			pbABI.AmountLimit = append(pbABI.AmountLimit, toPbAmountLimit(al))
		}
		ret.Abis = append(ret.Abis, pbABI)
	}
	return ret
}

func toCoreTx(t *rpcpb.TransactionRequest) *tx.Tx {
	ret := &tx.Tx{
		Time:       t.Time,
		Expiration: t.Expiration,
		GasPrice:   int64(t.GasPrice * 100),
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
	for _, s := range t.Signatures {
		ret.Signs = append(ret.Signs, &crypto.Signature{
			Algorithm: crypto.Algorithm(s.Algorithm),
			Pubkey:    s.PublicKey,
			Sig:       s.Signature,
		})
	}
	for _, s := range t.PublisherSigs {
		ret.PublishSigns = append(ret.PublishSigns, &crypto.Signature{
			Algorithm: crypto.Algorithm(s.Algorithm),
			Pubkey:    s.PublicKey,
			Sig:       s.Signature,
		})
	}
	return ret
}
