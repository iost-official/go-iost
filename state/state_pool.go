package state

import (
	"github.com/iost-official/prototype/core"
	"fmt"
)

type Pool interface {
	Copy() Pool
	GetPatch() Patch
	Flush() error

	Put(key Key, value Value) error
	Get(key Key) (Value, error)
	Has(key Key) (bool, error)
	Delete(key Key) error
}

func NewPool(bc core.BlockChain) Pool {
	return nil
}

type PoolImpl struct {
	db     Database
	patch  Patch
	parent *PoolImpl
}

func (p *PoolImpl) Copy() Pool {
	return nil
}
func (p *PoolImpl) GetPatch() Patch {
	return p.patch
}

func (p *PoolImpl) Put(key Key, value Value) error {
	exist, err := p.Has(key)
	if err != nil {
		return err
	}
	if exist {
		old, err := p.Get(key)
		if err != nil {
			return err
		}
		ans, err := Diff(old, value)
		if err != nil {
			return err
		}
		p.patch.Put(key, ans)
	} else {
		p.patch.Put(key, value)
	}
	return nil
}

func (p *PoolImpl) Get(key Key) (Value, error) {
	var val1 Value
	var err error
	if p.parent == nil {
		val1, err = p.db.Get(key)
		if err != nil {
			return nil, err
		}
	}
	val2, err := p.patch.Get(key)
	if err != nil {
		return nil, err
	}
	return Merge(val1, val2)

}
func (p *PoolImpl) Has(key Key) (bool, error) {
	val, err := p.patch.Get(key)
	if err != nil {
		return false, err
	}
	if val == nil {
		if p.parent != nil {
			return p.parent.Has(key)
		} else {
			return p.db.Has(key)
		}
	} else {
		if val.Type() == Nil {
			return false, nil
		} else {
			return true, nil
		}
	}

}
func (p *PoolImpl) Delete(key Key) error {
	if ok, _ := p.Has(key); ok {
		p.patch.Put(key, &ValueImpl{
			t: Nil,
		})
	} else {
		return fmt.Errorf("not found")
	}
	return nil
}

func (p *PoolImpl) Flush() error{
	p.parent.Flush()
	for k, v := range p.patch.m {
		if v.Type() != Nil {
			val0, err := p.db.Get(k)
			val, err := Merge(val0,v)
			if err != nil {
				return err
			}
			p.db.Put(k, val)
		} else {
			p.db.Delete(k)
		}
	}
	return nil
}
