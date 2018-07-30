package new_vm

import (
	"github.com/hashicorp/golang-lru"
)

type Pool struct {
	cache *lru.Cache
}

func NewPool(length int) (*Pool, error) {
	pool := &Pool{}
	var err error
	pool.cache, err = lru.New(length)
	return pool, err
}

func (v * Pool) Contract(contractName string) (*Contract, error) {
	contract, ok := v.cache.Get(contractName)
	if !ok {
		// todo 从数据库中找到contract，并且调用cache.Add
	}
	return contract.(*Contract), nil
}

func (v * Pool) DeleteContract(contractName string) error {
	return nil
}

