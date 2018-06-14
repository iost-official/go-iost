package state

import (
	"fmt"

	"sync"

	"github.com/iost-official/prototype/db"
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
	if val2 == VNil {
		return val1, nil
	}
	return Merge(val1, val2), nil
}
func (p *PoolImpl) Has(key Key) bool {
	ok := p.patch.Has(key)
	if ok {
		val := p.patch.Get(key)
		if val == VDelete {
			return false
		} else if val != VNil {
			return true
		}
	}

	if p.parent != nil {
		return p.parent.Has(key)
	} else {
		ok, _ := p.db.Has(key)
		return ok
	}

}
func (p *PoolImpl) Delete(key Key) {
	p.patch.Put(key, VDelete)
}

func (p *PoolImpl) Flush() error {
	if p.parent != nil {
		p.parent.Flush()
	}
	for k, v := range p.patch.m {
		switch {
		case v == VDelete:
			p.db.Delete(k)
		case v.Type() == Map:
			vm := v.(*VMap)
			for f, v := range vm.m {
				v0, err := p.db.GetHM(k, f)
				if err != nil {
					return err
				}
				val := Merge(v0, v)
				p.db.PutHM(k, f, val)
			}
		default:
			val0, err := p.db.Get(k)
			if err != nil {
				return err
			}
			val := Merge(val0, v)

			p.db.Put(k, val)
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
			fmt.Println(err) // TODO 消灭多余输出
			val1 = VNil
		}
	} else {
		val1, err = p.parent.GetHM(key, field)
		if err != nil {
			fmt.Println(err) // todo 消灭多余输出
			val1 = VNil
		}
		//fmt.Println("in GetHM get parent: ", key, field, val1)
	}

	val2 := p.patch.Get(key)
	if val2 == VNil {
		return val1, nil
	} else {
		if val2.Type() != Map {
			return nil, fmt.Errorf("type error : %v is not a hashmap", key)
		}
		val3 := val2.(*VMap).Get(field)

		//fmt.Println("in gethm :", val1, val3)
		//fmt.Println(Merge(val1, val3))

		return Merge(val1, val3), nil
	}
}
func (p *PoolImpl) PutHM(key, field Key, value Value) error {
	//fmt.Println("call put hm", key, field, value)
	//fmt.Println("len:", p.patch.Length())
	//fmt.Println("parent:", p.parent)
	if ok := p.patch.Has(key); ok {
		m := p.patch.Get(key)
		//fmt.Println("1.8", m)
		if m.Type() == Map {
			m.(*VMap).Set(field, value)
			//fmt.Println("1.9", m)
			p.patch.Put(key, m)
			return nil
		}
	}
	//fmt.Println("new map")
	m := MakeVMap(nil)
	m.Set(field, value)
	p.patch.Put(key, m)
	return nil
}

func (p *PoolImpl) MergeParent() (Pool, error) {
	bak := *(p.parent)
	for k, v := range p.patch.m {
		switch {
		case v == VDelete:
			bak.Delete(k)
		case v.Type() == Map:
			vm := v.(*VMap)
			for f, v := range vm.m {
				v0, err := bak.GetHM(k, f)
				if err != nil {
					return nil, err
				}
				val := Merge(v0, v)
				bak.PutHM(k, f, val)
			}
		default:
			val0, err := bak.Get(k)
			if err != nil {
				return nil, err
			}
			val := Merge(val0, v)
			bak.Put(k, val)
		}
	}
	return &bak, nil
}

func (p *PoolImpl) Parent() *PoolImpl {
	return p.parent
}

var StdPool Pool

var once sync.Once

func PoolInstance() error {
	bdb, err := db.DatabaseFactory("redis")
	if err != nil {
		return err
	}
	mdb := NewDatabase(bdb)
	if StdPool == nil {
		once.Do(func() {
			StdPool = NewPool(mdb)
		})
	}

	return nil
}
