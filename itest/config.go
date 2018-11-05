package itest

import (
	"os"

	"github.com/iost-official/go-iost/ilog"
	"gopkg.in/yaml.v2"
)

type ITestConfig struct {
	iserver map[string]string
	bank    *Account
}

func NewITestConfig(configfile string) *ITestConfig {
	file, err := os.Open(configfile)
	if err != nil {
		ilog.Fatalf("Open config file failed: %v", err)
	}

	data := []byte{}
	if _, err := file.Read(data); err != nil {
		ilog.Fatalf("Read config file failed: %v", err)
	}

	itc := &ITestConfig{}
	if err := yaml.Unmarshal(data, itc); err != nil {
		ilog.Fatalf("Unmarshal config file failed: %v", err)
	}

	return itc
}
