package verifier

import (
	"fmt"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/vm"
)

const (
	MaxBlockGas uint64 = 1000000
)

//go:generate gencode go -schema=structs.schema -package=verifier

// 底层verifier，用来组织vm，不要直接使用
type Verifier struct {
	Pool state.Pool
	VMMonitor
}

func (v *Verifier) Verify(contract vm.Contract) (state.Pool, uint64, error) {
	_, pool, gas, err := v.Call(v.Pool, contract.Info().Prefix, "main")
	return pool, gas, err
}

func (v *Verifier) SetPool(pool state.Pool) {
	v.Pool = pool
}

// 验证新tx的工具类
type CacheVerifier struct {
	Verifier
}

// 验证contract，返回pool是包含了该contract的pool。如果contain为true则进行合并
//
// 取得tx中的Contract的方法： tx.Contract
func (cv *CacheVerifier) VerifyContract(contract vm.Contract, contain bool) (state.Pool, error) {
	sender := contract.Info().Sender
	var balanceOfSender float64
	val0, err := cv.Pool.GetHM("iost", state.Key(sender))
	val, ok := val0.(*state.VFloat)
	if !ok {
		return nil, fmt.Errorf("type error")
	}
	balanceOfSender = val.ToFloat64()

	if balanceOfSender < float64(contract.Info().GasLimit)*contract.Info().Price {
		return nil, fmt.Errorf("balance not enough")
	}

	cv.StartVM(contract)
	pool, gas, err := cv.Verify(contract)
	if err != nil {
		cv.StopVm(contract)
		return nil, err
	}
	cv.StopVm(contract)

	if gas > uint64(contract.Info().GasLimit) {
		balanceOfSender -= float64(contract.Info().GasLimit) * contract.Info().Price
		val1 := state.MakeVFloat(balanceOfSender)
		cv.Pool.PutHM("iost", state.Key(sender), val1)
		return nil, fmt.Errorf("gas exceeded")
	}

	balanceOfSender -= float64(gas) * contract.Info().Price
	if balanceOfSender < 0 {
		balanceOfSender = 0
		val1 := state.MakeVFloat(balanceOfSender)
		cv.Pool.PutHM("iost", state.Key(sender), val1)
		return nil, fmt.Errorf("can not afford gas")
	}
	val1 := state.MakeVFloat(balanceOfSender)
	cv.Pool.PutHM("iost", state.Key(sender), val1)

	if contain {
		cv.SetPool(pool)
	}

	return pool, nil
}

func NewCacheVerifier(pool state.Pool) CacheVerifier {
	cv := CacheVerifier{
		Verifier: Verifier{
			Pool:      pool.Copy(),
			VMMonitor: NewVMMonitor(),
		},
	}
	return cv
}

func (cv *CacheVerifier) Close() {
	cv.Verifier.Stop()
}

// 验证block的工具类
type BlockVerifier struct {
	CacheVerifier
	oldPool state.Pool
}

// 验证block，返回pool是包含了该block的pool。如果contain为true则进行合并
func (bv *BlockVerifier) VerifyBlock(b block.Block, contain bool) (state.Pool, error) {
	bv.oldPool = bv.Pool
	for i := 0; i < b.TxLen(); i++ {
		c := b.TxGet(i).Contract
		_, err := bv.VerifyContract(c, true)
		if err != nil {
			return nil, err
		}

	}
	if contain {
		return bv.Pool, nil
	}
	newPool := bv.Pool
	bv.Pool = bv.oldPool
	return newPool, nil
}

func NewBlockVerifier(pool state.Pool) BlockVerifier {
	bv := BlockVerifier{
		CacheVerifier: NewCacheVerifier(pool),
	}
	return bv
}
