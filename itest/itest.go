package itest

import (
	"fmt"
	"math/rand"
)

type ITest struct {
	bank    *Account
	keys    []*Key
	clients []*Client
}

func New(itc *ITestConfig, keys []*Key) *ITest {
	return &ITest{
		bank:    itc.Bank,
		keys:    keys,
		clients: itc.Clients,
	}
}

func (t *ITest) CreateAccount(name string) (*Account, error) {
	if len(t.keys) == 0 {
		return nil, fmt.Errorf("keys is empty")
	}
	if len(t.clients) == 0 {
		return nil, fmt.Errorf("clients is empty")
	}
	kIndex := rand.Intn(len(t.keys))
	key := t.keys[kIndex]
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	account, err := client.CreateAccount(t.bank, name, key)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (t *ITest) Transfer(sender *Account, token, recipient, amount string) error {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	err := client.Transfer(sender, token, recipient, amount)
	if err != nil {
		return err
	}

	return nil
}

func (t *ITest) SetContract(contract *Contract) error {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	err := client.SetContract(t.bank, contract)
	if err != nil {
		return err
	}

	return nil
}

func (t *ITest) GetTransaction(hash string) (*Transaction, error) {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	transaction, err := client.GetTransaction(hash)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (t *ITest) GetAccount(name string) (*Account, error) {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	account, err := client.GetAccount(name)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (t *ITest) SendTransaction(transaction *Transaction) (string, error) {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	hash, err := client.SendTransaction(transaction)
	if err != nil {
		return "", err
	}

	return hash, nil
}
