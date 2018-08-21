package database

import (
	"os"

	"io/ioutil"

	"encoding/json"

	"strings"

	"fmt"

	"errors"
	"github.com/bitly/go-simplejson"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
)

type database struct {
	json *simplejson.Json
}

func NewDatabase() *database {
	var data database
	data.json = simplejson.New()
	data.addSystem()
	return &data
}

func NewDatabaseFromPath(path string) *database {
	var data database
	json, err := readJson(path)
	if err != nil {
		return nil
	}
	data.json = json
	return &data
}

func (d *database) Get(table string, key string) (string, error) {
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
func (d *database) Put(table string, key string, value string) error {
	if strings.HasPrefix(key, "c-") {
		d.json.Set(key, value)
	} else {
		d.json.Set(key, MustUnmarshal(value))
	}
	return nil
}
func (d *database) Del(table string, key string) error {
	d.json.Del(key)
	return nil
}
func (d *database) Has(table string, key string) (bool, error) {
	_, ok := d.json.CheckGet(key)
	return ok, nil
}
func (d *database) Keys(table string, prefix string) ([]string, error) {
	return nil, nil
}
func (d *database) Save(path string) error {

	d.json.Del("c-iost.system")

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	buf, err := d.json.EncodePretty()
	if err != nil {
		return err
	}
	f.Write(buf)
	f.Close()
	return nil
}
func (d *database) Load(path string) error {
	var err error
	d.json, err = readJson(path)
	if err != nil {
		return err
	}
	d.addSystem()
	return nil
}

func LoadBlockhead(path string) (*block.BlockHead, error) {

	json, err := readJson(path)
	if err != nil {
		return nil, err
	}
	bh := &block.BlockHead{}
	bh.Time, err = json.Get("time").Int64()
	bh.ParentHash, err = json.Get("parent_hash").Bytes()
	bh.Number, err = json.Get("number").Int64()
	bh.Witness, err = json.Get("witness").String()
	return bh, err

}

func LoadTxInfo(path string) (*tx.Tx, error) {

	json, err := readJson(path)
	if err != nil {
		return nil, err
	}

	t := &tx.Tx{}
	s := make([][]byte, 0)
	for _, v := range json.Get("signers").MustArray() {
		s = append(s, []byte(v.(string)))
	}

	t.Signers = s
	p, err := json.Get("publisher").String()
	t.Publisher.Pubkey = account.GetPubkeyByID(p)
	t.GasLimit, err = json.Get("gas_limit").Int64()
	t.GasPrice, err = json.Get("gas_price").Int64()
	return t, err
}

func readJson(path string) (*simplejson.Json, error) {
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

func (d *database) addSystem() {
	cp := contract.Compiler{}

	f, err := os.Open("system.json")
	if err != nil {
		panic(err)
	}
	fd, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	c, err := cp.Parse("iost.system", "", string(fd))
	if err != nil {
		panic(err)
	}
	fmt.Println(c)

	d.json.Set("c-iost.system", c.Encode())
}

func (d *database) Commit() {

}

func (d *database) Rollback() {

}
