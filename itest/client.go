package itest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/rpc/pb"
	"google.golang.org/grpc"
)

// Constant of Client
var (
	Interval   = 15 * time.Second
	Timeout    = (90 + 30) * time.Second
	InitToken  = "iost"
	InitAmount = "1000000"
	InitPledge = "1000000"
	InitRAM    = "1000000"
)

// Error of Client
var (
	ErrTimeout = fmt.Errorf("Transaction be on chain timeout")
)

// Client is a grpc client for iserver
type Client struct {
	grpc  rpcpb.ApiServiceClient
	mutex sync.Mutex
	Name  string
	Addr  string
}

func (c *Client) getGRPC() (rpcpb.ApiServiceClient, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.grpc == nil {
		conn, err := grpc.Dial(c.Addr, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		c.grpc = rpcpb.NewApiServiceClient(conn)
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
		&rpcpb.TxHashRequest{
			Hash: hash,
		},
	)
	if err != nil {
		return nil, err
	}

	return NewTransactionFromPb(resp.Transaction), nil
}

// GetReceipt will get receipt by tx hash
func (c *Client) GetReceipt(hash string) (*Receipt, error) {
	grpc, err := c.getGRPC()
	if err != nil {
		return nil, err
	}

	resp, err := grpc.GetTxReceiptByTxHash(
		context.Background(),
		&rpcpb.TxHashRequest{
			Hash: hash,
		},
	)
	if err != nil {
		return nil, err
	}

	return NewReceiptFromPb(resp), nil
}

// GetAccount will get account by name
func (c *Client) GetAccount(name string) (*Account, error) {
	grpc, err := c.getGRPC()
	if err != nil {
		return nil, err
	}

	resp, err := grpc.GetAccount(
		context.Background(),
		&rpcpb.GetAccountRequest{
			Name:           name,
			ByLongestChain: true,
		},
	)
	if err != nil {
		return nil, err
	}

	// TODO: Get account permission by resp
	account := &Account{
		ID:      name,
		balance: resp.GetBalance(),
	}

	return account, nil
}

// SendTransaction will send transaction to blockchain
func (c *Client) SendTransaction(transaction *Transaction) (string, error) {
	grpc, err := c.getGRPC()
	if err != nil {
		return "", err
	}

	resp, err := grpc.SendTransaction(
		context.Background(),
		transaction.ToTxRequest(),
	)
	if err != nil {
		return "", err
	}

	ilog.Debugf("Check transaction receipt for %v...", resp.GetHash())
	if err := c.checkTransaction(resp.GetHash()); err != nil {
		return "", err
	}
	ilog.Debugf("Check transaction receipt for %v successful!", resp.GetHash())

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

			ilog.Debugf("Get receipt for %v...", hash)
			r, err := c.GetReceipt(hash)
			if err != nil {
				break
			}
			ilog.Debugf("Get receipt for %v successful!", hash)

			if !r.Success() {
				return fmt.Errorf("%v: %v", r.Status.Code, r.Status.Message)
			}
			return nil
		}
	}
}

// CreateAccount will create account by sending transaction
func (c *Client) CreateAccount(creator *Account, name string, key *Key) (*Account, error) {
	action1 := tx.NewAction(
		"auth.iost",
		"SignUp",
		fmt.Sprintf(`["%v", "%v", "%v"]`, name, key.ID, key.ID),
	)

	action2 := tx.NewAction(
		"ram.iost",
		"buy",
		fmt.Sprintf(`["%v", "%v", %v]`, creator.ID, name, InitRAM),
	)

	action3 := tx.NewAction(
		"gas.iost",
		"pledge",
		fmt.Sprintf(`["%v", "%v", "%v"]`, creator.ID, name, InitPledge),
	)

	action4 := tx.NewAction(
		"token.iost",
		"transfer",
		fmt.Sprintf(`["%v", "%v", "%v", "%v", ""]`, InitToken, creator.ID, name, InitAmount),
	)

	actions := []*tx.Action{action1, action2, action3, action4}
	transaction := NewTransaction(actions)

	st, err := creator.Sign(transaction)
	if err != nil {
		return nil, err
	}

	ilog.Debugf("Sending create account transaction for %v...", name)
	if _, err := c.SendTransaction(st); err != nil {
		return nil, err
	}
	ilog.Debugf("Sended create account transaction for %v!", name)

	account := &Account{
		ID:      name,
		balance: InitAmount,
		key:     key,
	}

	return account, nil
}

// Transfer will transfer token by sending transaction
func (c *Client) Transfer(sender, recipient *Account, token, amount string) error {
	action := tx.NewAction(
		"token.iost",
		"transfer",
		fmt.Sprintf(`["%v", "%v", "%v", "%v", ""]`, token, sender.ID, recipient.ID, amount),
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
		"system.iost",
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
