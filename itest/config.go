package itest

import (
	"encoding/json"
	"io/ioutil"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
)

// Constant of itest config
const (
	DefaultITestConfig = `
{
  "bank":{
    "id": "admin",
    "seckey": "2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1",
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
	ilog.Infof("Load itest config from file...")

	var data []byte
	if file == "" {
		data = []byte(DefaultITestConfig)
	} else {
		var err error
		data, err = ioutil.ReadFile(file)
		if err != nil {
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
