package itest

import "github.com/iost-official/go-iost/core/contract"

type Contract struct {
	*contract.Contract
}

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

func (c *Contract) String() string {
	return c.B64Encode()
}
