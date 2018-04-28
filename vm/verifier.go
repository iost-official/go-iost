package vm

import (
	"github.com/iost-official/prototype/state"
	"runtime"
	"fmt"
)

type Verifier struct {
	Pool   state.Pool
	Prefix string
	Vms    map[string]VM
}

func (v *Verifier) StartVm(contract Contract) {
	switch contract.(type) {
	case *LuaContract:
		var lvm LuaVM
		lvm.Prepare(contract.(*LuaContract), v.Pool, v.Prefix)
		lvm.Start()
		v.Vms[string(contract.Hash())] = &lvm
	}
}
func (v *Verifier) StopVm(contract Contract) {
	v.Vms[string(contract.Hash())].Stop()
	delete(v.Vms, string(contract.Hash()))
}
func (v *Verifier) Stop() {
	for _, vm := range v.Vms {
		vm.Stop()
	}
}

func (v *Verifier) Verify(contract Contract) (state.Pool, error) {

	vm, ok := v.Vms[string(contract.Hash())]
	if !ok {
		return nil, fmt.Errorf("not prepared")
	}

	_, pool, err := vm.Call("main")
	return pool, err
}

func (v *Verifier) SetPool(pool state.Pool) {
	v.Pool = pool
	for _, vm := range v.Vms {
		vm.SetPool(pool)
	}
}

type CacheVerifier struct {
	Verifier
}

func (cv *CacheVerifier) VerifyContract(contract Contract, contain bool) (state.Pool, error) {
	cv.StartVm(contract)
	pool, err := cv.Verify(contract)
	if err != nil {
		return nil, err
	}
	if contain {
		cv.SetPool(pool)
	}
	cv.StopVm(contract)
	return cv.Pool, nil
}

func NewCacheVerifier(pool state.Pool) CacheVerifier {
	cv := CacheVerifier{
		Verifier: Verifier{
			Pool:   pool.Copy(),
			Prefix: "cache+",
			Vms:    make(map[string]VM),
		},
	}
	runtime.SetFinalizer(cv, func() {
		cv.Verifier.Stop()
	})
	return cv
}

func VerifyBlock(contracts []Contract, pool state.Pool) (state.Pool, error) {
	cv := NewCacheVerifier(pool)
	for _, c := range contracts {
		_, err := cv.VerifyContract(c, true)
		if err != nil {
			return nil, err
		}
	}
	return cv.Pool, nil
}
