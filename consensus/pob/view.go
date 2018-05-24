package pob

import (
	"time"

	"github.com/iost-official/prototype/account"
)

type WitnessView interface {
	IsPrimary(account2 account.Account) bool
	IsWitness(account2 account.Account) bool
	NextDue() time.Duration
}

type View struct {
	Primary account.Account
	Backups []account.Account
}

func (v *View) IsPrimary(account2 account.Account) bool {
	return v.Primary.ID == account2.ID
}

func (v *View) IsWitness(account2 account.Account) bool {
	if v.Primary.ID == account2.ID {
		return true
	} else {
		for _, a := range v.Backups {
			if a.ID == account2.ID {
				return true
			}
		}
	}
	return false
}
func (v *View) NextDue() time.Duration {
	return 10000
}
