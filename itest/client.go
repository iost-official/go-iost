package itest

import (
	"context"
	"fmt"
	"time"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/rpc"
	"google.golang.org/grpc"
)

// Constant of Client
var (
	Interval = 15 * time.Second
	Timeout  = (90 + 30) * time.Second
)

// Error of Client
var (
	ErrTimeout = fmt.Errorf("Transaction be on chain timeout")
)

type Client struct {
	grpc rpc.ApisClient
	name string
	addr string
}

func (c *Client) getGRPC() (rpc.ApisClient, error) {
	if c.grpc == nil {
		conn, err := grpc.Dial(c.addr)
		if err != nil {
			return nil, err
		}
		c.grpc = rpc.NewApisClient(conn)
		ilog.Infof("Create grpc connection with %v successful", c.addr)
	}
	return c.grpc, nil
}

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
		Tx: (&tx.Tx{}).FromPb(resp.GetTxRaw()),
	}

	return transaction, nil
}

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
		TxReceipt: (&tx.TxReceipt{}).FromPb(resp.GetTxReceiptRaw()),
	}

	return receipt, nil
}

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

func (c *Client) SendTransaction(transaction *Transaction) (string, error) {
	grpc, err := c.getGRPC()
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
				if r.Success() {
					return nil
				} else {
					return fmt.Errorf("%v: %v", r.Status.Code, r.Status.Message)
				}
			}
		}
	}
}

func (c *Client) CreateAccount(name string, key *Key, creater *Account) (*Account, error) {
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

	actions := []*tx.Action{action1, action2}
	transaction := NewTransaction(actions)

	st, err := creater.Sign(transaction)
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

func (c *Client) Transfer(token, sender, recipient, amount string) error {
	return nil
}

func (c *Client) SetContract(abi, code string) error {
	return nil
}
