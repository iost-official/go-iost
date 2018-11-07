package itest

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/rpc"
)

type ITest struct {
	bank    *Account
	keys    []*Key
	clients []*Client
}

func New(bank *Account, keys []*Key, clients []*Client) *ITest {
	return &ITest{
		bank:    bank,
		keys:    keys,
		clients: clients,
	}
}

func (t *ITest) CreateAccount(name string) (*Account, error) {
	index := rand.Intn(len(t.keys))
	key := t.keys[index]

	action1 := tx.NewAction(
		"iost.auth",
		"SignUp",
		fmt.Sprintf(
			`["%v", "%v", "%v"]`,
			name,
			key.ID,
			key.ID,
		),
	)

	action2 := tx.NewAction(
		"iost.auth",
		"SignUp",
		fmt.Sprintf(
			`["%v", "%v", "%v"]`,
			name,
			key.ID,
			key.ID,
		),
	)

	actions := []*tx.Action{&action1, &action2}
	signers := []string{t.bank.ID}
	transaction := NewTransaction(actions, signers)
	_, err := t.SendTransaction(transaction)
	if err != nil {
		return nil, err
	}

	// TODO: Check transaction by hash

	return t.GetAccount(name)
}

func (t *ITest) GetAccount(name string) (*Account, error) {
	index := rand.Intn(len(t.clients))
	grpc, err := t.clients[index].GetGRPC()
	if err != nil {
		return nil, err
	}

	_, err := grpc.GetAccountInfo(
		context.Background(),
		&rpc.GetAccountReq{
			ID:              name,
			UseLongestChain: true,
		},
	)
	if err != nil {
		return nil, err
	}

	// TODO: Check account by resp

	return nil, nil
}

func (t *ITest) Transfer(token, sender, recipient, amount string) error {
	return nil
}

func (t *ITest) SetContract(abi, code string) error {
	return nil
}

func (t *ITest) SendTransaction(transaction *Transaction) (string, error) {
	index := rand.Intn(len(t.clients))
	grpc, err := t.clients[index].GetGRPC()
	if err != nil {
		return "", err
	}

	resp, err := grpc.SendRawTx(
		context.Background(),
		&rpc.RawTxReq{
			Data: transaction.Encode(),
		},
	)
	if err != nil {
		return "", err
	}

	return resp.GetHash(), nil
}
