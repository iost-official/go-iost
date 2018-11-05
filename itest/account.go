package itest

import "github.com/iost-official/go-iost/account"

type Account struct {
	ID  string
	key *account.KeyPair
}
