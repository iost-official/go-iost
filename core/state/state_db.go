package state

import (
	"fmt"
	"github.com/iost-official/prototype/db"
)

// db的适配器
type Database struct {
	db db.Database
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
			d.db.Put(key.Encode(), []byte(value.String()))
		}
		for k, v := range vi.m {
			d.db.PutHM(key.Encode(), k.Encode(), []byte(v.String()))
		}
	}
	return d.db.Put(key.Encode(), []byte(value.String()))
}
func (d *Database) Get(key Key) (Value, error) {
	raw, err := d.db.Get(key.Encode())
	if err != nil {
		return nil, err
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
	} else if len(raw) < 1 {
		return nil, fmt.Errorf("not found")
	}
	return ParseValue(string(raw[0]))
}
func (d *Database) PutHM(key, field Key, value Value) error {
	return d.db.PutHM(key.Encode(), field.Encode(), []byte(value.String()))
}
