package state

import "github.com/iost-official/prototype/iostdb"

type Database struct {
	db iostdb.Database
}

func NewDatabase(db iostdb.Database) Database {
	return Database{
		db: db,
	}
}

func(d *Database) Put(key Key, value Value) error{
	return d.db.Put(key.Encode(), value.Encode())
}
func(d *Database) Get(key Key) (Value, error){
	var v Value
	raw, err := d.db.Get(key.Encode())
	v.Decode(raw)
	return v, err
}
func(d *Database) Has(key Key) (bool, error){
	return d.db.Has(key.Encode())
}
func(d *Database) Delete(key Key) error{
	return d.db.Delete(key.Encode())
}