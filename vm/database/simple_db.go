package database

import (
	"os"

	"io/ioutil"

	"encoding/json"

	"strings"

	"fmt"

	"errors"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
)

// SimpleDB implements simple database interface
type SimpleDB struct {
	json *simplejson.Json
}

// NewDatabase returns a SimpleDB with empty data and system contract
func NewDatabase() *SimpleDB {
	var data SimpleDB
	data.json = simplejson.New()
	return &data
}

// NewDatabaseFromPath returns a SimpleDB with data loaded from json file
func NewDatabaseFromPath(path string) *SimpleDB {
	var data SimpleDB
	json, err := readJSON(path)
	if err != nil {
		return nil
	}
	data.json = json
	return &data
}

// Get key-value from db with marshal
func (d *SimpleDB) Get(table string, key string) (string, error) {
	jso := d.json.Get(key)
	var out interface{}
	var err error

	if jso.Interface() == nil {
		out = nil
	} else {
		switch jso.Interface().(type) {
		case json.Number:
			out, err = jso.Int64()
		case string:
			out, err = jso.String()
		case bool:
			out, err = jso.Bool()
		}
	}
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(key, "c-") {
		if out == nil {
			return "", errors.New("key not exists")
		}
		return out.(string), nil
	}

	return Marshal(out)
}

// Put key-value into db with unmarshal
func (d *SimpleDB) Put(table string, key string, value string) error {
	if strings.HasPrefix(key, "c-") {
		d.json.Set(key, value)
	} else {
		d.json.Set(key, MustUnmarshal(value))
	}
	return nil
}

// Del delete key from db
func (d *SimpleDB) Del(table string, key string) error {
	d.json.Del(key)
	return nil
}

// Has return if key exists in db
func (d *SimpleDB) Has(table string, key string) (bool, error) {
	_, ok := d.json.CheckGet(key)
	return ok, nil
}

// Keys do nothing
func (d *SimpleDB) Keys(table string, prefix string) ([]string, error) {
	return nil, nil
}

// Save save db data to json file
func (d *SimpleDB) Save(path string) error {

	d.json.Del("c-system.iost")

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	buf, err := d.json.EncodePretty()
	if err != nil {
		return err
	}
	_, err = f.Write(buf)
	if err != nil {
		return err
	}
	err = f.Close()
	return err
}

// Load load db data from json file
func (d *SimpleDB) Load(path string) error {
	var err error
	d.json, err = readJSON(path)
	if err != nil {
		return err
	}
	return nil
}

// LoadBlockhead load block info as block.BlockHead from json file
func LoadBlockhead(path string) (*block.BlockHead, error) {

	json, err := readJSON(path)
	if err != nil {
		return nil, err
	}
	bh := &block.BlockHead{}
	bh.Time, err = json.Get("time").Int64()
	if err != nil {
		return nil, err
	}
	bh.ParentHash, err = json.Get("parent_hash").Bytes()
	if err != nil {
		return nil, err
	}
	bh.Number, err = json.Get("number").Int64()
	if err != nil {
		return nil, err
	}
	bh.Witness, err = json.Get("witness").String()
	if err != nil {
		return nil, err
	}
	return bh, nil

}

// LoadTxInfo load tx info as tx.Tx from json file
func LoadTxInfo(path string) (*tx.Tx, error) {

	json, err := readJSON(path)
	if err != nil {
		return nil, err
	}

	t := &tx.Tx{}
	s := make([]string, 0)
	for _, v := range json.Get("signers").MustArray() {
		s = append(s, v.(string))
	}

	t.Signers = s
	p, err := json.Get("publisher").String()
	if err != nil {
		return nil, err
	}
	t.PublishSigns = append(t.PublishSigns, &crypto.Signature{
		Pubkey: account.GetPubkeyByID(p),
	})
	t.GasLimit, err = json.Get("gas_limit").Int64()
	if err != nil {
		return nil, err
	}
	t.GasPrice, err = json.Get("gas_price").Int64()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func readJSON(path string) (*simplejson.Json, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fd, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return simplejson.NewJson(fd)
}

// AddSystem load system contract and data from json file
func (d *SimpleDB) AddSystem(path string) {
	cp := contract.Compiler{}

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	fd, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	c, err := cp.Parse("system.iost", "", string(fd))
	if err != nil {
		panic(err)
	}
	fmt.Println(c)

	d.json.Set("c-system.iost", c.Encode())
}

// Commit do nothing
func (d *SimpleDB) Commit() {

}

// Rollback do nothing
func (d *SimpleDB) Rollback() {

}
