package common

import (
	"os"

	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

type DBConfig struct {
	LdbPath string
}

type VMConfig struct {
}

type P2PConfig struct {
	Address string
	Port    int64
}

// Config provide all configuration for the application
type Config struct {
	VM  *VMConfig
	DB  *DBConfig
	P2P *P2PConfig
}

// NewConfig returns a new instance of Config
func NewConfig(configfile string) *Config {
	v := viper.GetViper()
	v.SetConfigType("yaml")

	f, err := os.Open(configfile)
	if err != nil {
		ilog.Fatal("Failed to open config file '%v', %v", configfile, err)
	}

	if err := v.ReadConfig(f); err != nil {
		ilog.Fatal("Failed to read config file: %v", err)
	}

	c := &Config{}
	if err := v.Unmarshal(c); err != nil {
		ilog.Fatal("Unable to decode into struct, %v", err)
	}

	return c
}

func (c *Config) YamlString() string {
	bs, err := yaml.Marshal(c)
	if err != nil {
		ilog.Fatal("Unable to marshal config to YAML: %v", err)
	}
	return string(bs)
}
