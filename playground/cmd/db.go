package cmd

import (
	"strings"
)

type Database struct {
	Normal map[string][]byte
}

func (d *Database) Put(key []byte, value []byte) error {
	d.Normal[string(key)] = value
	return nil
}
func (d *Database) PutHM(key []byte, args ...[]byte) error {
	key1 := string(key)
	key2 := string(args[0])
	d.Normal[key1+"."+key2] = args[1]
	return nil
}
func (d *Database) Get(key []byte) ([]byte, error) {
	return d.Normal[string(key)], nil
}
func (d *Database) GetHM(key []byte, args ...[]byte) ([][]byte, error) {
	//fmt.Println("GetHM:", string(key), string(args[0]))
	key1 := string(key)
	key2 := string(args[0])
	return [][]byte{d.Normal[key1+"."+key2]}, nil
}
func (d *Database) Has(key []byte) (bool, error) {
	_, ok := d.Normal[string(key)]
	return ok, nil
}
func (d *Database) Delete(key []byte) error {
	delete(d.Normal, string(key))
	return nil
}
func (d *Database) Close() {
}
func (d *Database) Type(key string) (string, error) {
	for k := range d.Normal {
		if strings.HasPrefix(k, key+".") {
			return "hash", nil
		}
	}
	return "string", nil
}
func (d *Database) GetAll(key string) (map[string]string, error) {
	mm := make(map[string]string)
	for k, v := range d.Normal {
		if strings.HasPrefix(k, key+".") {
			f := k[strings.Index(k, "."):]
			mm[f] = string(v)
		}
	}
	return mm, nil
}
