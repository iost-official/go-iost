package itest

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/iost-official/go-iost/ilog"
)

// Constant of itest
const (
	Zero = 1e-6
)

// ITest is the test controller
type ITest struct {
	bank    *Account
	keys    []*Key
	clients []*Client
}

// New will return the itest by config and keys
func New(c *Config, keys []*Key) *ITest {
	return &ITest{
		bank:    c.Bank,
		keys:    keys,
		clients: c.Clients,
	}
}

// Load will load the itest from file
func Load(keysfile, configfile string) (*ITest, error) {
	ilog.Infof("Load itest from file...")

	keys, err := LoadKeys(keysfile)
	if err != nil {
		return nil, fmt.Errorf("load keys failed: %v", err)
	}

	itc, err := LoadConfig(configfile)
	if err != nil {
		return nil, fmt.Errorf("load itest config failed: %v", err)
	}

	it := New(itc, keys)

	ilog.Infof("Load itest from file successful!")
	return it, nil
}

// CreateAccountN will create n accounts concurrently
func (t *ITest) CreateAccountN(num int) ([]*Account, error) {
	ilog.Infof("Create %v account...", num)

	res := make(chan interface{})
	for i := 0; i < num; i++ {
		go func(n int, res chan interface{}) {
			name := fmt.Sprintf("account%04d", n)
			account, err := t.CreateAccount(name)
			if err != nil {
				res <- err
			} else {
				res <- account
			}
		}(i, res)
	}

	accounts := []*Account{}
	for i := 0; i < num; i++ {
		switch value := (<-res).(type) {
		case error:
			ilog.Errorf("Create account failed: %v", value)
		case *Account:
			accounts = append(accounts, value)
		default:
			return nil, fmt.Errorf("unexpect res: %v", value)
		}
	}

	if len(accounts) != num {
		return nil, fmt.Errorf(
			"expect create %v account, but only created %v account",
			num,
			len(accounts),
		)
	}

	ilog.Infof("Create %v account successful!", len(accounts))

	// TODO Get account by rpc, and compare account result

	return accounts, nil
}

// CreateAccount will create a account by name
func (t *ITest) CreateAccount(name string) (*Account, error) {
	if len(t.keys) == 0 {
		return nil, fmt.Errorf("keys is empty")
	}
	if len(t.clients) == 0 {
		return nil, fmt.Errorf("clients is empty")
	}
	kIndex := rand.Intn(len(t.keys)) // nolint: golint
	key := t.keys[kIndex]
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	account, err := client.CreateAccount(t.bank, name, key)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// TransferN will send n transfer transaction concurrently
func (t *ITest) TransferN(num int, accounts []*Account) error {
	ilog.Infof("Send %v transfer transaction...", num)

	res := make(chan interface{})
	for i := 0; i < num; i++ {
		go func(res chan interface{}) {
			A := accounts[rand.Intn(len(accounts))]
			B := accounts[rand.Intn(len(accounts))]
			amount := float64(rand.Int63n(10000)+1) / 100

			A.AddBalance(-amount)
			B.AddBalance(amount)
			ilog.Debugf("Transfer %v -> %v, amount: %v", A.ID, B.ID, fmt.Sprintf("%0.8f", amount))

			res <- t.Transfer(A, B, "iost", fmt.Sprintf("%0.8f", amount))
		}(res)
	}

	for i := 0; i < num; i++ {
		switch value := (<-res).(type) {
		case error:
			return fmt.Errorf("Send transfer transaction failed: %v", value)
		default:
		}
	}

	ilog.Infof("Send %v transfer transaction successful!", num)

	return nil
}

// ContractTransferN will send n contract transfer transaction concurrently
func (t *ITest) ContractTransferN(cid string, num int, accounts []*Account) error {
	ilog.Infof("Send %v contract transfer transaction...", num)

	res := make(chan interface{})
	for i := 0; i < num; i++ {
		go func(res chan interface{}) {
			A := accounts[rand.Intn(len(accounts))]
			B := accounts[rand.Intn(len(accounts))]
			amount := float64(rand.Int63n(10000)+1) / 100

			A.AddBalance(-amount)
			B.AddBalance(amount)
			ilog.Debugf("Contract transfer %v -> %v, amount: %v", A.ID, B.ID, fmt.Sprintf("%0.8f", amount))

			res <- t.ContractTransfer(cid, A, B, fmt.Sprintf("%0.8f", amount))
		}(res)
	}

	for i := 0; i < num; i++ {
		switch value := (<-res).(type) {
		case error:
			return fmt.Errorf("Send contract transfer transaction failed: %v", value)
		default:
		}
	}

	ilog.Infof("Send %v contract transfer transaction successful!", num)

	return nil
}

// CheckAccounts will check account info by getting account info
func (t *ITest) CheckAccounts(a []*Account) error {
	ilog.Infof("Get %v accounts info...", len(a))

	res := make(chan interface{})
	for _, i := range a {
		go func(name string, res chan interface{}) {
			account, err := t.GetAccount(name)
			if err != nil {
				res <- err
			} else {
				res <- account
			}
		}(i.ID, res)
	}

	aMap := make(map[string]*Account)
	for i := 0; i < len(a); i++ {
		switch value := (<-res).(type) {
		case error:
			ilog.Errorf("Get account failed: %v", value)
		case *Account:
			aMap[value.ID] = value
		default:
			return fmt.Errorf("unexpect res: %v", value)
		}
	}

	if len(aMap) != len(a) {
		return fmt.Errorf(
			"expect get %v account, but only ge %v account",
			len(a),
			len(aMap),
		)
	}

	ilog.Infof("Get %v accounts info successful!", len(aMap))

	ilog.Infof("Check %v accounts info...", len(a))

	for _, i := range a {
		expect := i.Balance()
		actual := aMap[i.ID].Balance()
		if math.Abs(expect-actual) > Zero {
			return fmt.Errorf(
				"expect account %v's balance is %0.8f, but balance is %0.8f",
				i.ID,
				expect,
				actual,
			)
		}
	}
	ilog.Infof("Check %v accounts info successful!", len(a))

	return nil
}

// ContractTransfer will contract transfer token from sender to recipient
func (t *ITest) ContractTransfer(cid string, sender, recipient *Account, amount string) error {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	err := client.ContractTransfer(cid, sender, recipient, amount)
	if err != nil {
		return err
	}

	return nil
}

// Transfer will transfer token from sender to recipient
func (t *ITest) Transfer(sender, recipient *Account, token, amount string) error {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	err := client.Transfer(sender, recipient, token, amount)
	if err != nil {
		return err
	}

	return nil
}

// SetContract will set the contract on blockchain
func (t *ITest) SetContract(contract *Contract) (string, error) {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	hash, err := client.SetContract(t.bank, contract)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// GetTransaction will get transaction by tx hash
func (t *ITest) GetTransaction(hash string) (*Transaction, error) {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	transaction, err := client.GetTransaction(hash)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetAccount will get account by name
func (t *ITest) GetAccount(name string) (*Account, error) {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	account, err := client.GetAccount(name)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// SendTransaction will send transaction to blockchain
func (t *ITest) SendTransaction(transaction *Transaction) (string, error) {
	cIndex := rand.Intn(len(t.clients))
	client := t.clients[cIndex]

	hash, err := client.SendTransaction(transaction)
	if err != nil {
		return "", err
	}

	return hash, nil
}
