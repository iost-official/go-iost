/*
Package verifier, adapter and monitor of vm to IOST node
*/
package verifier

import (
	"fmt"

	"reflect"

	"regexp"

	"errors"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
)

const (
	MaxBlockGas uint64  = 1000000
	TxBaseFee   float64 = 0.01
)

//go:generate gencode go -schema=structs.schema -package=verifier

type Verifier struct {
	vmMonitor
	Context *vm.Context
}

func (v *Verifier) Verify(contract vm.Contract, pool state.Pool) (state.Pool, uint64, error) {
	_, err := v.RestartVM(contract)
	if err != nil {
		return pool, 0, err
	}
	_, pool, gas, err := v.Call(v.Context, pool, contract.Info().Prefix, "main")
	return pool, gas, err
}

type CacheVerifier struct {
	Verifier
}

func balanceOfSender(sender vm.IOSTAccount, pool state.Pool) float64 {
	val0, err := pool.GetHM("iost", state.Key(sender))
	if err != nil {
		return 0
	}
	val, ok := val0.(*state.VFloat)
	if val0 == state.VNil {
		return 0
	} else if !ok {
		panic(fmt.Errorf("pool type error: should VFloat, acture %v; in iost.%v",
			reflect.TypeOf(val0).String(), string(sender)))
	}
	return val.ToFloat64()
}

func setBalanceOfSender(sender vm.IOSTAccount, pool state.Pool, amount float64) {
	pool.PutHM("iost", state.Key(sender), state.MakeVFloat(amount))
}

func (cv *CacheVerifier) VerifyContract(contract vm.Contract, pool state.Pool) (state.Pool, error) {
	if contract.Info().Price < 0 {
		return pool, errors.New("illegal gas price")
	}

	sender := contract.Info().Publisher
	bos := balanceOfSender(sender, pool)
	if bos < float64(contract.Info().GasLimit)*contract.Info().Price+TxBaseFee {
		return pool, fmt.Errorf("balance not enough: sender:%v balance:%f\n", string(sender), bos)
	}

	_, err := cv.RestartVM(contract)
	if err != nil {
		return pool, err
	}
	pool, gas, err := cv.Verify(contract, pool)
	if err != nil {
		return pool, err
	}

	bos2 := balanceOfSender(sender, pool)

	if gas > uint64(contract.Info().GasLimit) {
		return pool, errors.New("gas overflow")
	}

	bos2 -= float64(gas)*contract.Info().Price + TxBaseFee
	if bos2 < 0 {
		return pool, fmt.Errorf("can not afford gas")
	}

	setBalanceOfSender(sender, pool, bos2)
	return pool, nil
}

func NewCacheVerifier() CacheVerifier {
	cv := CacheVerifier{
		Verifier: Verifier{
			vmMonitor: newVMMonitor(),
		},
	}
	return cv
}

func (cv *CacheVerifier) CleanUp() {
	cv.Verifier.Stop()
}

func ParseGenesis(c vm.Contract, pool state.Pool) (state.Pool, error) {
	cachePool := pool.Copy()
	code := c.Code()
	rePutHM := regexp.MustCompile(`@PutHM[\t ]*([^\t ]*)[\t ]*([^\t ]*)[\t ]*([^\n\t ]*)[\n\t ]*`)
	rePut := regexp.MustCompile(`@Put[\t ]+([^\t ]*)[\t ]*([^\n\t ]*)[\n\t ]*`)
	allHM := rePutHM.FindAllStringSubmatch(code, -1)
	allPut := rePut.FindAllStringSubmatch(code, -1)
	for _, hm := range allHM {
		v, err := state.ParseValue(hm[3])
		if err != nil {
			panic(err)
		}
		cachePool.PutHM(state.Key(hm[1]), state.Key(hm[2]), v)
	}
	for _, put := range allPut {
		v, err := state.ParseValue(put[2])
		if err != nil {
			panic(err)
		}
		cachePool.Put(state.Key(put[1]), v)
	}
	return cachePool, nil
}
