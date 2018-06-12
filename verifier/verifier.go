package verifier

import (
	"fmt"

	"reflect"

	"regexp"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
)

const (
	MaxBlockGas uint64 = 1000000
)

//go:generate gencode go -schema=structs.schema -package=verifier

// 底层verifier，用来组织vm，不要直接使用
type Verifier struct {
	vmMonitor
}

func (v *Verifier) Verify(contract vm.Contract, pool state.Pool) (state.Pool, uint64, error) {
	v.RestartVM(contract)
	//fmt.Println(v.Pool.GetHM("iost", "b"))
	_, pool, gas, err := v.Call(pool, contract.Info().Prefix, "main")
	//fmt.Println(pool.GetHM("iost", "b"))
	return pool, gas, err
}

// 验证新tx的工具类
type CacheVerifier struct {
	Verifier
}

//// 验证contract，返回pool是包含了该contract的pool。如果contain为true则进行合并
////
//// 取得tx中的Contract的方法： tx.Contract
//func (cv *CacheVerifier) VerifyContract(contract vm.Contract, contain bool) (state.Pool, error) {
//	sender := contract.Info().Publisher
//	var balanceOfSender float64
//	val0, err := cv.Pool.GetHM("iost", state.Key(sender))
//	if err != nil {
//		return nil, err
//	}
//	val, ok := val0.(*state.VFloat)
//	if val0 == state.VNil {
//		val = state.MakeVFloat(0)
//	} else if !ok {
//		return nil, fmt.Errorf("pool type error: should VFloat, acture %v; in iost.%v",
//			reflect.TypeOf(val0).String(), string(sender))
//	}
//	balanceOfSender = val.ToFloat64()
//
//	if balanceOfSender < float64(contract.Info().GasLimit)*contract.Info().Price {
//		return nil, fmt.Errorf("balance not enough")
//	}
//
//	cv.StartVM(contract)
//	pool, gas, err := cv.Verify(contract)
//	if err != nil {
//		cv.StopVM(contract)
//		return nil, err
//	}
//	cv.StopVM(contract)
//
//	if contract.Info().Price == 0 {
//		if contain {
//			cv.SetPool(pool)
//		}
//		return pool, nil
//	}
//
//	val1, err := pool.GetHM("iost", state.Key(sender))
//	if err != nil {
//		return nil, err
//	}
//
//	if gas > uint64(contract.Info().GasLimit) {
//		balanceOfSender -= float64(contract.Info().GasLimit) * contract.Info().Price
//		if balanceOfSender < 0 {
//			panic("balance of sender to be zero!!")
//		}
//		val1 := state.MakeVFloat(balanceOfSender)
//		pool2 := cv.Pool.Copy()
//		pool2.PutHM("iost", state.Key(sender), val1)
//		if contain {
//			cv.SetPool(pool)
//		}
//		return pool2, nil
//	}
//
//	balanceOfSender = val1.(*state.VFloat).ToFloat64()
//
//	balanceOfSender -= float64(gas) * contract.Info().Price
//	if balanceOfSender < 0 {
//		balanceOfSender = 0
//		val1 := state.MakeVFloat(balanceOfSender)
//		pool2 := cv.Pool.Copy()
//		pool2.PutHM("iost", state.Key(sender), val1)
//		if contain {
//			cv.SetPool(pool)
//		}
//		return nil, fmt.Errorf("can not afford gas")
//	}
//
//	val2 := state.MakeVFloat(balanceOfSender)
//	pool.PutHM("iost", state.Key(sender), val2)
//
//	if contain {
//		cv.SetPool(pool)
//	}
//	return pool, nil
//}

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
// 取得tx中的Contract的方法： tx.Contract
func (cv *CacheVerifier) VerifyContract(contract vm.Contract, pool state.Pool) (state.Pool, error) {
	sender := contract.Info().Publisher
	if contract.Info().Price != 0 {
		bos := balanceOfSender(sender, pool)
		if bos < float64(contract.Info().GasLimit)*contract.Info().Price {
			return pool, fmt.Errorf("balance not enough:balance %f gaslimit*price:%f", bos, float64(contract.Info().GasLimit)*contract.Info().Price)
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
			panic("gas overflow!")
			//bos -= float64(contract.Info().GasLimit) * contract.Info().Price
			//if bos < 0 {
			//	panic("balance of sender to be zero!!")
			//}
			//val1 := state.MakeVFloat(bos)
			//pool.PutHM("iost", state.Key(sender), val1)
			//return pool, nil
		}

		bos2 -= float64(gas) * contract.Info().Price
		if bos2 < 0 {
			//bos -= float64(gas) * contract.Info().Price
			//val1 := state.MakeVFloat(bos)
			//pool2 := pool.Copy()
			//pool2.PutHM("iost", state.Key(sender), val1)
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

// 验证block的工具类
//type BlockVerifier struct {
//	CacheVerifier
//	oldPool state.Pool
//}
//
//// 验证block，返回pool是包含了该block的pool。如果contain为true则进行合并
//func (bv *BlockVerifier) VerifyBlock(b *block.Block, contain bool) (state.Pool, error) {
//	bv.oldPool = bv.Pool
//	for i := 0; i < b.LenTx(); i++ {
//		c := b.GetTx(i).Contract
//		_, err := bv.VerifyContract(c, true)
//		if err != nil {
//			return nil, err
//		}
//
//	}
//	if contain {
//		return bv.Pool, nil
//	}
//	newPool := bv.Pool
//	bv.Pool = bv.oldPool
//	return newPool, nil
//}
//
//func NewBlockVerifier(pool state.Pool) BlockVerifier {
//	bv := BlockVerifier{
//		CacheVerifier: NewCacheVerifier(pool),
//	}
//	return bv
//}
//
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
