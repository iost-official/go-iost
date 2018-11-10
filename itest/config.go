package itest

import (
	"encoding/json"
	"os"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
)

const (
	DefaultITestConfig = `
{
  "bank":{
    "id": "admin",
    "seckey": "zRGrECumZAomHRJa1Jr9u4HdGypeBDJGvyF7XEcjh5cJcK7aBGBPWF5MWf9NsfjtgqSYnrXBZmEyUZ8NzSJ4LVT",
    "algorithm":"ed25519"
  },
  "clients":[
    {
      "name": "iserver",
      "addr": "127.0.0.1:30002"
    }
  ]
}
`
)

type ITestConfig struct {
	Bank    *Account
	Clients []*Client
}

func LoadITestConfig(file string) (*ITestConfig, error) {
	data := []byte{}
	if file == "" {
		data = []byte(DefaultITestConfig)
	} else {

		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}

		if _, err := f.Read(data); err != nil {
			return nil, err
		}
	}

	itc := &ITestConfig{}
	if err := json.Unmarshal(data, itc); err != nil {
		return nil, err
	}

	ilog.Debugf("Bank id: %v", itc.Bank.ID)
	ilog.Debugf("Bank seckey: %v", common.Base58Encode(itc.Bank.key.Seckey))
	ilog.Debugf("Clients name: %v", itc.Clients[0].Name)
	ilog.Debugf("Clients addr: %v", itc.Clients[0].Addr)

	return itc, nil
}
