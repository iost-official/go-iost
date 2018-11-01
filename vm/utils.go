package vm

import (
	"errors"
	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
)

// CheckPublisher check publisher of tx
func CheckPublisher(db database.IMultiValue, t *tx.Tx) error {
	auth := map[string]int{account.GetIDByPubkey(t.PublishSign.Pubkey): 2}
	reenter := make(map[string]int)
	vi := database.NewVisitor(0, db)
	ok, _ := host.Auth(vi, t.Publisher, "active", auth, reenter)
	if !ok {
		return errors.New("missing authority")
	}
	return nil
}

// CheckSigners check signers of tx
func CheckSigners(db database.IMultiValue, t *tx.Tx) error {

	auth := make(map[string]int)

	for _, v := range t.Signs {
		keyname := account.GetIDByPubkey(v.Pubkey)
		auth[keyname] = 1
	}
	auth[account.GetIDByPubkey(t.PublishSign.Pubkey)] = 2

	vi := database.NewVisitor(100, db)
	for _, a := range t.Signers {
		x := strings.Split(a, "@")
		if len(x) != 2 {
			return errors.New("illegal permission")
		}
		reenter := make(map[string]int)
		ok, _ := host.Auth(vi, x[0], x[1], auth, reenter)
		if !ok {
			return errors.New("missing authority")
		}
	}
	return nil
}
