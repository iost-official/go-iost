package itest

type ITest struct {
	bank    *Account
	keys    []*Key
	iserver map[string]string
}

func New(bank *Account, keys []*Key, iserver map[string]string) *ITest {
	return &ITest{
		bank:    bank,
		keys:    keys,
		iserver: iserver,
	}
}

func (t *ITest) CreateAccount(name string) (*Account, error) {
	return nil, nil
}

func (t *ITest) Transfer(token, sender, recipient, amount string) error {
	return nil
}

func (t *ITest) SetContract(abi, code string) error {
	return nil
}

func (t *ITest) SendTransaction(tx *Transaction) error {
	return nil
}
