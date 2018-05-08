package state

import (
	"fmt"
)

//go:generate gencode go -schema=structs.schema -package=state
//go:generate mockgen -destination ../mocks/mock_pool.go -package core_mock github.com/iost-official/prototype/core/state Pool

type PoolImpl struct {
	db     Database
	patch  Patch
	parent *PoolImpl
}

// 通过一个db生成新的pool
func NewPool(db Database) Pool {
	return &PoolImpl{
		db:     db,
		patch:  Patch{make(map[Key]Value)},
		parent: nil,
	}
}

func (p *PoolImpl) Copy() Pool {
	pp := PoolImpl{
		db:     p.db,
		patch:  Patch{make(map[Key]Value)},
		parent: p,
	}
	return &pp
}
func (p *PoolImpl) GetPatch() Patch {
	return p.patch
}

func (p *PoolImpl) Put(key Key, value Value) {
	p.patch.Put(key, value)
}

func (p *PoolImpl) Get(key Key) (Value, error) {
	var val1 Value
	var err error
	if p.parent == nil {
		val1, err = p.db.Get(key)
		if err != nil {
			val1 = VNil
		}
	} else {
		val1, err = p.parent.Get(key)
		if err != nil {
			val1 = VNil
		}
	}
	val2 := p.patch.Get(key)
	return Merge(val1, val2)
}
func (p *PoolImpl) Has(key Key) bool {
	ok := p.patch.Has(key)
	if !ok {
		if p.parent != nil {
			return p.parent.Has(key)
		} else {
			ok, _ := p.db.Has(key)
			return ok
		}
	} else {
		val := p.patch.Get(key)
		if val == VNil {
			return false
		} else {
			return true
		}
	}

}
func (p *PoolImpl) Delete(key Key) {
	p.patch.Put(key, VNil)
}

func (p *PoolImpl) Flush() error {
	p.parent.Flush()
	for k, v := range p.patch.m {
		if v.Type() != Nil {
			val0, err := p.db.Get(k)
			val, err := Merge(val0, v)
			if err != nil {
				return err
			}
			p.db.Put(k, val)
		} else {
			p.db.Delete(k)
		}
	}
	p.parent = nil
	return nil
}

func (p *PoolImpl) GetHM(key, field Key) (Value, error) {
	var err error

	var val1 Value
	if p.parent == nil {
		val1, err = p.db.GetHM(key, field)
		if err != nil {
			val1 = VNil
		}
	} else {
		val1, err = p.parent.GetHM(key, field)
		if err != nil {
			val1 = VNil
		}
	}

	val2 := p.patch.Get(key)
	if val2 == VNil {
		return val1, nil
	} else {
		if val2.Type() != Map {
			return nil, fmt.Errorf("type error")
		}
		val3, err := val2.(*VMap).Get(field)
		if err != nil {
			return nil, err
		}

		return Merge(val1, val3)
	}
}
func (p *PoolImpl) PutHM(key, field Key, value Value) error {
	if ok := p.patch.Has(key); ok {
		m := p.patch.Get(key)

		if m.Type() == Map {
			m.(*VMap).Set(field, value)
		}
	}
	m := MakeVMap(nil)
	m.Set(field, value)
	p.patch.Put(key, m)
	return nil
}
