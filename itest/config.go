package itest

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ITestConfig struct {
	iserver map[string]string
	bank    *Account
}

func NewITestConfig(configfile string) *ITestConfig {
	file, err := os.Open(configfile)
	if err != nil {
		log.Fatalf("Open config file failed: %v", err)
	}

	data := []byte{}
	if _, err := file.Read(data); err != nil {
		log.Fatalf("Read config file failed: %v", err)
	}

	itc := &ITestConfig{}
	if err := yaml.Unmarshal(data, itc); err != nil {
		log.Fatalf("Unmarshal config file failed: %v", err)
	}

	return itc
}
