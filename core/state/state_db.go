package state

import (
	"github.com/iost-official/prototype/db"
)

type Database struct {
	db db.Database
}

type HashDatabase interface {
	db.Database
	Type(key string) (string, error)
	GetAll(key string) (map[string]string, error)
}

func NewDatabase(db db.Database) Database {
	return Database{
		db: db,
	}
}

func (d *Database) Put(key Key, value Value) error {
	switch value.Type() {
	case Map:
		vi, ok := value.(*VMap)
		if !ok {
			d.db.Put(key.Encode(), []byte(value.EncodeString()))
		}
		for k, v := range vi.m {
			d.db.PutHM(key.Encode(), k.Encode(), []byte(v.EncodeString()))
		}
	}
	return d.db.Put(key.Encode(), []byte(value.EncodeString()))
}
func (d *Database) Get(key Key) (Value, error) {
	rdb, ok := d.db.(HashDatabase)
	if ok {
		t, err := rdb.Type(string(key))
		if err != nil {
			return nil, err
		}
		//fmt.Println(t)
		if t == "hash" {
			ms, err := rdb.GetAll(string(key))
			if err != nil {
				return nil, err
			}
			m := MakeVMap(nil)
			for k, v := range ms {
				val, err := ParseValue(v)
				if err != nil {
					return nil, err
				}
				m.Set(Key(k), val)
			}
			return m, nil
		}
	}
	raw, err := d.db.Get(key.Encode())
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return VNil, nil
	}
	return ParseValue(string(raw))
}
func (d *Database) Has(key Key) (bool, error) {
	return d.db.Has(key.Encode())
}
func (d *Database) Delete(key Key) error {
	return d.db.Delete(key.Encode())
}
func (d *Database) GetHM(key, field Key) (Value, error) {
	raw, err := d.db.GetHM(key.Encode(), field.Encode())
	if err != nil {
		return nil, err
	}
	if raw == nil || raw[0] == nil {
		return VNil, nil
	}
	return ParseValue(string(raw[0]))
}
func (d *Database) PutHM(key, field Key, value Value) error {
	return d.db.PutHM(key.Encode(), field.Encode(), []byte(value.EncodeString()))
}
