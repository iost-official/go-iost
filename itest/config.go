package itest

import (
	"encoding/json"
	"os"
)

type ITestConfig struct {
	bank    *Account
	clients []*Client
}

func LoadITestConfig(file string) (*ITestConfig, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	data := []byte{}
	if _, err := f.Read(data); err != nil {
		return nil, err
	}

	itc := &ITestConfig{}
	if err := json.Unmarshal(data, itc); err != nil {
		return nil, err
	}

	return itc, nil
}
