package itest

import (
	"encoding/json"
	"os"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
)

// Constant of itest config
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

// Config is the config of itest
type Config struct {
	Bank    *Account
	Clients []*Client
}

// LoadConfig will load the itest config from file
func LoadConfig(file string) (*Config, error) {
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

	c := &Config{}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}

	ilog.Debugf("Bank id: %v", c.Bank.ID)
	ilog.Debugf("Bank seckey: %v", common.Base58Encode(c.Bank.key.Seckey))
	ilog.Debugf("Clients name: %v", c.Clients[0].Name)
	ilog.Debugf("Clients addr: %v", c.Clients[0].Addr)

	return c, nil
}
