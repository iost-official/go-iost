package verifier

import (
	"fmt"

	"reflect"

	"regexp"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
	"github.com/pkg/errors"
)

const (
	MaxBlockGas uint64 = 1000000
)

//go:generate gencode go -schema=structs.schema -package=verifier

// 底层verifier，用来组织vm，不要直接使用
type Verifier struct {
	vmMonitor
	Context *vm.Context
}

func (v *Verifier) Verify(contract vm.Contract, pool state.Pool) (state.Pool, uint64, error) {
	v.RestartVM(contract)
	//fmt.Println(v.Pool.GetHM("iost", "b"))
	_, pool, gas, err := v.Call(v.Context, pool, contract.Info().Prefix, "main")
	//fmt.Println(pool.GetHM("iost", "b"))
	return pool, gas, err
}

// 验证新tx的工具类
type CacheVerifier struct {
	Verifier
}

func balanceOfSender(sender vm.IOSTAccount, pool state.Pool) float64 {
	val0, err := pool.GetHM("iost", state.Key(sender))
	if err != nil {
		panic(err)
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

// 验证contract，返回pool是包含了该contract的pool。传入时要注意是传入pool还是pool.copy()
//
// 取得tx中的Contract的方法： tx.contract
func (cv *CacheVerifier) VerifyContract(contract vm.Contract, pool state.Pool) (state.Pool, error) {
	sender := contract.Info().Publisher
	if contract.Info().Price != 0 {
		bos := balanceOfSender(sender, pool)
		if bos < float64(contract.Info().GasLimit)*contract.Info().Price {
			return pool, fmt.Errorf("balance not enough: sender:%v balance:%f\n", string(sender), bos)
		}
	}

	cv.RestartVM(contract)
	pool, gas, err := cv.Verify(contract, pool)
	if err != nil {
		//cv.StopVM(contract)
		return pool, err
	}
	//cv.StopVM(contract)

	if contract.Info().Price != 0 {

		bos2 := balanceOfSender(sender, pool)

		if gas > uint64(contract.Info().GasLimit) { // TODO 不应该发生的分支
			return pool, errors.New("gas overflow!")
		}

		bos2 -= float64(gas) * contract.Info().Price
		if bos2 < 0 {
			return pool, fmt.Errorf("can not afford gas")
		}

		setBalanceOfSender(sender, pool, bos2)
	}
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
	// TODO 应在这里初始化一个全新的state pool
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
