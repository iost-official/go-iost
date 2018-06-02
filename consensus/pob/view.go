package pob

import (
	"time"

	"github.com/iost-official/prototype/account"
)

type WitnessView interface {
	IsPrimary(account2 account.Account) bool
	IsWitness(account2 account.Account) bool
	IsBackup(account2 account.Account) bool
	NextDue() time.Duration
	Next()
}

var Due = time.Duration(24) * time.Hour
var Period, WitnessSize, Order time.Duration

type View struct {
	Primary   account.Account
	Witness   []account.Account
	Backups   []account.Account
	BuildTime time.Duration
}

func (v *View) IsPrimary(account2 account.Account) bool {
	return v.Primary.ID == account2.ID
}
func (v *View) IsWitness(account2 account.Account) bool {
	if v.Primary.ID == account2.ID {
		return true
	} else {
		for _, a := range v.Witness {
			if a.ID == account2.ID {
				return true
			}
		}
	}
	return false
}
func (v *View) IsBackup(account2 account.Account) bool {
	for _, a := range v.Backups {
		if a.ID == account2.ID {
			return true
		}
	}
	return false
}
func (v *View) NextDue() time.Duration {
	now := time.Duration(time.Now().Unix()) * time.Second
	return v.BuildTime + Due - now
}
func (v *View) Next() {
	now := time.Duration(time.Now().Unix()) * time.Second
	if now > v.BuildTime+Due {
		ac := account.Account{Pubkey: account.GetPubkeyByID(string(StdServiPool.BestUser().owner))}
		ac.ID = string(StdServiPool.BestUser().owner)
		if !v.IsWitness(ac) {
			v.Primary = ac
		}
	}
	v.Witness = append(v.Witness, v.Primary)
	v.Primary = v.Witness[0]
	v.Witness = v.Witness[1:]
}
func InitialView() View {
	return View{}
}
