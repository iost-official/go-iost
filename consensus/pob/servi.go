package pob

import "github.com/iost-official/prototype/account"

type Servi struct {
}

type ServiPool struct {
}

func (sp *ServiPool) Pop() account.Account {
	return account.Account{}
}
