package vm

import (
	"errors"
	"fmt"
	"github.com/iost-official/go-iost/vm/native"
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

// CheckTxGasLimitValid ...
func CheckTxGasLimitValid(t *tx.Tx, currentGas *common.Fixed, dbVisitor *database.Visitor) (err error) {
	gasLimit := &common.Fixed{Value: t.GasLimit, Decimal: 2}
	if !currentGas.LessThan(gasLimit) {
		return nil
	}
	defaultErr := fmt.Errorf("gas not enough: user %v has %v < %v", t.Publisher, currentGas.ToString(), gasLimit.ToString())
	if !(len(t.Actions) == 1 && t.Actions[0].Contract == native.GasContractName && t.Actions[0].ActionName == "pledge") {
		return defaultErr
	}
	// user is trying to pledge for gas without initial gas
	args, err := UnmarshalArgs(dbVisitor.Contract(native.GasContractName).ABI("pledge"), t.Actions[0].Data)
	if err != nil {
		return fmt.Errorf("invalid gas pledge args %v %v", err, t.Actions[0].Data)
	}
	if !(args[0] == t.Publisher && args[1] == t.Publisher) {
		return defaultErr
	}
	balance := dbVisitor.TokenBalanceFixed("iost", t.Publisher)
	pledgeAmount, err := common.NewFixed(args[2].(string), 8)
	if err != nil {
		return fmt.Errorf("invalid gas pledge amount %v %v", err, args[2].(string))
	}
	if pledgeAmount.LessThan(native.GasMinPledgeOfUser) {
		return fmt.Errorf("invalid gas pledge amount %v %v", err, args[2].(string))
	}
	if balance.LessThan(pledgeAmount) {
		return fmt.Errorf("iost token amount not enough for pledgement %v < %v", balance.ToString(), pledgeAmount.ToString())
	}
	if currentGas.Add(pledgeAmount.Multiply(native.GasImmediateReward)).LessThan(gasLimit) {
		return fmt.Errorf("gas not enough even if considering the new gas pledgement %v + %v < %v",
			currentGas.ToString(), pledgeAmount.Multiply(native.GasImmediateReward).ToString(), gasLimit.ToString())
	}
	return nil
}
