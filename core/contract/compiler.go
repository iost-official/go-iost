package contract

import (
	"github.com/bitly/go-simplejson"
)

/*
{
"lang": "javascript",
"version": "1.0.0",

"abi": [{
	"name": "abc",
	"args": ["", "", ""],
	"payment": 0,
	"cost_limit": [1,2,3], // data net cpu
	"price_limit": 123
}
*/

type Compiler struct {
}

func (c *Compiler) parseABI(json *simplejson.Json) ([]*ABI, error) {

	abis := make([]*ABI, 0)
	array, err := json.Array()
	if err != nil {
		return nil, err
	}
	for i := range array {

		var abi ABI

		ja := json.GetIndex(i)
		func() {
			defer func() {
				if err0 := recover(); err0 != nil {
					err = err0.(error)
				}
			}()

			abi.Name = ja.Get("name").MustString()
			abi.Args = ja.Get("args").MustStringArray()
			if _, ok := ja.CheckGet("payment"); ok {
				abi.Payment = int32(ja.Get("payment").MustInt())

				data := int64(ja.Get("cost_limit").GetIndex(0).MustInt())
				net := int64(ja.Get("cost_limit").GetIndex(1).MustInt())
				cpu := int64(ja.Get("cost_limit").GetIndex(2).MustInt())
				abi.Limit = NewCost(data, net, cpu)
				abi.GasPrice = ja.Get("price_limit").MustInt64()
			} else {
				abi.Payment = 0
				abi.Limit = NewCost(1, 1, 1)
				abi.GasPrice = 1
			}

		}()
		if err != nil {
			return nil, err
		}

		abis = append(abis, &abi)

	}
	return abis, nil
}

func (c *Compiler) parseInfo(json *simplejson.Json) (*Info, error) {
	var info Info
	var err error
	info.Lang, err = json.Get("lang").String()
	if err != nil {
		return nil, err
	}
	info.VersionCode, err = json.Get("version").String()
	if err != nil {
		return nil, err
	}
	info.Abis, err = c.parseABI(json.Get("abi"))
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *Compiler) Parse(id, code, abi string) (*Contract, error) {
	json, err := simplejson.NewJson([]byte(abi))
	if err != nil {
		return nil, err
	}

	info, err := c.parseInfo(json)
	if err != nil {
		return nil, err
	}

	var con Contract
	con.Info = info
	con.Code = code
	con.ID = id
	return &con, nil
}
