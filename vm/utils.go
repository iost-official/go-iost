package vm

import (
	"errors"
	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
)

// CheckPublisher check publisher of tx
func CheckPublisher(db database.IMultiValue, t *tx.Tx) error {
	auth := map[string]int{}
	for _, sig := range t.PublishSigns {
		auth[account.GetIDByPubkey(sig.Pubkey)] = 2
	}
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
	for _, sig := range t.PublishSigns {
		auth[account.GetIDByPubkey(sig.Pubkey)] = 2
	}

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

// CheckAmountLimit check amountLimit of tx valid
func CheckAmountLimit(db database.IMultiValue, t *tx.Tx) error {
	vi := database.NewVisitor(100, db)
	for _, limit := range t.AmountLimit {
		decimal := vi.Decimal(limit.Token)
		if decimal == -1 {
			return errors.New("token not exists in amountLimit")
		}
		_, err := common.NewFixed(limit.Val, decimal)
		if err != nil {
			return err
		}
	}
	return nil
}
