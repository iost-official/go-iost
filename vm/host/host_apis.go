package host

import (
	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/vm"
	"github.com/pkg/errors"
)

var l log.Logger

var (
	ErrBalanceNotEnough = errors.New("balance not enough")
)

func Put(pool state.Pool, key state.Key, value state.Value) bool {
	pool.Put(key, value)
	return true
}

func Get(pool state.Pool, key state.Key) (state.Value, error) {
	return pool.Get(key)
}

func Log(s, cid string) {
	l.D("From Lua %v > %v", cid, s)
}

func Transfer(pool state.Pool, src, des string, value float64) bool {

	err := changeToken(pool, "iost", state.Key(src), -value)

	if err != nil {
		return false
	}

	err = changeToken(pool, "iost", state.Key(des), value)
	if err != nil {
		return false
	}

	return true
}

func Deposit(pool state.Pool, contractPrefix, payer string, value float64) bool {
	err := changeToken(pool, "iost", state.Key(payer), -value)
	if err != nil {
		return false
	}

	err = changeToken(pool, "iost-contract", state.Key(contractPrefix), value)
	if err != nil {
		return false
	}

	return true

}

func Withdraw(pool state.Pool, contractPrefix, payer string, value float64) bool {
	err := changeToken(pool, "iost-contract", state.Key(contractPrefix), -value)
	if err != nil {
		return false
	}
	err = changeToken(pool, "iost", state.Key(payer), value)
	if err != nil {
		return false
	}

	return true

}
func RandomByParentHash(ctx vm.Context, probability float64) bool {
	seed := ctx.ParentHash()

	return float64(common.Sha256(seed)[10]) > probability*255
}

func Publisher(contract vm.Contract) string {
	return string(contract.Info().Publisher)
}

func changeToken(pool state.Pool, key, field state.Key, delta float64) error {
	val0, err := pool.GetHM(state.Key(key), state.Key(field))
	if err != nil {
		return err
	}
	var val float64
	switch val0.(type) {
	case *state.VFloat:
		val = val0.(*state.VFloat).ToFloat64()
	default:
		val = 0
	}

	if val+delta < 0 {
		return ErrBalanceNotEnough
	}
	ba := state.MakeVFloat(val + delta)

	pool.PutHM(state.Key(key), state.Key(field), ba)
	return nil
}
