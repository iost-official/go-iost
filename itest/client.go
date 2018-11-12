package itest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/rpc"
	"google.golang.org/grpc"
)

// Constant of Client
var (
	Interval   = 15 * time.Second
	Timeout    = (90 + 30) * time.Second
	InitToken  = "IOST"
	InitAmount = "1000000"
)

// Error of Client
var (
	ErrTimeout = fmt.Errorf("Transaction be on chain timeout")
)

// Client is a grpc client for iserver
type Client struct {
	grpc  rpc.ApisClient
	mutex sync.Mutex
	Name  string
	Addr  string
}

func (c *Client) getGRPC() (rpc.ApisClient, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.grpc == nil {
		conn, err := grpc.Dial(c.Addr, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		c.grpc = rpc.NewApisClient(conn)
		ilog.Infof("Create grpc connection with %v successful", c.Addr)
	}
	return c.grpc, nil
}

// GetTransaction will get transaction by tx hash
func (c *Client) GetTransaction(hash string) (*Transaction, error) {
	grpc, err := c.getGRPC()
	if err != nil {
		return nil, err
	}

	resp, err := grpc.GetTxByHash(
		context.Background(),
		&rpc.HashReq{
			Hash: hash,
		},
	)
	if err != nil {
		return nil, err
	}

	transaction := &Transaction{
		Tx: (&tx.Tx{}).FromPb(resp.Tx),
	}

	return transaction, nil
}

// GetReceipt will get receipt by tx hash
func (c *Client) GetReceipt(hash string) (*Receipt, error) {
	grpc, err := c.getGRPC()
	if err != nil {
		return nil, err
	}

	resp, err := grpc.GetTxReceiptByTxHash(
		context.Background(),
		&rpc.HashReq{
			Hash: hash,
		},
	)
	if err != nil {
		return nil, err
	}

	receipt := &Receipt{
		TxReceipt: (&tx.TxReceipt{}).FromPb(resp.GetTxReceipt()),
	}

	return receipt, nil
}

// GetAccount will get account by name
func (c *Client) GetAccount(name string) (*Account, error) {
	grpc, err := c.getGRPC()
	if err != nil {
		return nil, err
	}

	resp, err := grpc.GetAccountInfo(
		context.Background(),
		&rpc.GetAccountReq{
			ID:              name,
			UseLongestChain: true,
		},
	)
	if err != nil {
		return nil, err
	}

	// TODO: Get account permission by resp
	account := &Account{
		ID:      name,
		Balance: resp.GetBalance(),
	}

	return account, nil
}

// SendTransaction will send transaction to blockchain
func (c *Client) SendTransaction(transaction *Transaction) (string, error) {
	grpc, err := c.getGRPC()
	if err != nil {
		return "", err
	}

	resp, err := grpc.SendTx(
		context.Background(),
		&rpc.TxReq{
			Tx: transaction.ToPb(),
		},
	)
	if err != nil {
		return "", err
	}

	if err := c.checkTransaction(resp.GetHash()); err != nil {
		return "", err
	}

	return resp.GetHash(), nil
}

func (c *Client) checkTransaction(hash string) error {
	ticker := time.NewTicker(Interval)
	afterTimeout := time.After(Timeout)
	for {
		select {
		case <-afterTimeout:
			return ErrTimeout
		default:
			<-ticker.C
			r, err := c.GetReceipt(hash)
			if err == nil && r != nil {
				if !r.Success() {
					return fmt.Errorf("%v: %v", r.Status.Code, r.Status.Message)
				}
				return nil
			}
		}
	}
}

// CreateAccount will create account by sending transaction
func (c *Client) CreateAccount(creator *Account, name string, key *Key) (*Account, error) {
	action1 := tx.NewAction(
		"iost.auth",
		"SignUp",
		fmt.Sprintf(`["%v", "%v", "%v"]`, name, key.ID, key.ID),
	)

	action2 := tx.NewAction(
		"iost.token",
		"transfer",
		fmt.Sprintf(`["%v", "%v", %v, %v]`, InitToken, creator.ID, name, InitAmount),
	)

	actions := []*tx.Action{action1, action2}
	transaction := NewTransaction(actions)

	st, err := creator.Sign(transaction)
	if err != nil {
		return nil, err
	}

	if _, err := c.SendTransaction(st); err != nil {
		return nil, err
	}

	account := &Account{
		ID:  name,
		key: key,
	}

	return account, nil
}

// Transfer will transfer token by sending transaction
func (c *Client) Transfer(sender *Account, token, recipient, amount string) error {
	action := tx.NewAction(
		"iost.token",
		"transfer",
		fmt.Sprintf(`["%v", "%v", %v, %v]`, token, sender.ID, recipient, amount),
	)

	actions := []*tx.Action{action}
	transaction := NewTransaction(actions)

	st, err := sender.Sign(transaction)
	if err != nil {
		return err
	}

	if _, err := c.SendTransaction(st); err != nil {
		return err
	}

	return nil
}

// SetContract will set the contract by sending transaction
func (c *Client) SetContract(creator *Account, contract *Contract) error {
	action := tx.NewAction(
		"iost.system",
		"SetCode",
		fmt.Sprintf(`["%v"]`, contract),
	)

	actions := []*tx.Action{action}
	transaction := NewTransaction(actions)

	st, err := creator.Sign(transaction)
	if err != nil {
		return err
	}

	if _, err := c.SendTransaction(st); err != nil {
		return err
	}

	return nil
}
