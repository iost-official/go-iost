package itest

import "github.com/iost-official/go-iost/core/contract"

// Contract is the contract object
type Contract struct {
	*contract.Contract
}

// NewContract will return a new contract
func NewContract(code, abi string) (*Contract, error) {
	c, err := (&contract.Compiler{}).Parse("", code, abi)
	if err != nil {
		return nil, err
	}

	contract := &Contract{
		Contract: c,
	}
	return contract, nil
}

// String will return the string of contract
func (c *Contract) String() string {
	return c.B64Encode()
}
