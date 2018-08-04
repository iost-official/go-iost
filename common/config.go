package common

import (
	"errors"
	"github.com/spf13/viper"
)

type Config struct {
	vip *viper.Viper

	LdbPath       string
	RedisAddr     string
	RedisPort     int64
	NetLogPath    string
	NodeTablePath string
	NodeID        string
	ListenAddr    string
	RegAddr       string
	RpcPort       string
	Target        string
	Port          int64
	MetricsPort   string
	AccSecKey     string
	LogPath       string
	CfgFile       string
	LogFile       string
	DbFile        string
}

func NewConfig(vip *viper.Viper) (*Config, error) {

	if vip == nil {
		return nil, errors.New("NewConfig vip error")
	}

	return &Config{vip: vip}, nil
}

func (c *Config) LocalConfig() error {

	c.CfgFile = c.vip.GetString("config")
	c.LogFile = c.vip.GetString("log")
	c.DbFile = c.vip.GetString("db")

	c.LogPath = c.vip.GetString("log.path")
	c.LdbPath = c.vip.GetString("ldb.path")
	c.RedisAddr = c.vip.GetString("redis.addr")
	c.RedisPort = c.vip.GetInt64("redis.port")
	c.NetLogPath = c.vip.GetString("net.log-path")
	c.NodeTablePath = c.vip.GetString("net.node-table-path")
	c.NodeID = c.vip.GetString("net.node-id") //optional
	c.ListenAddr = c.vip.GetString("net.listen-addr")
	c.RegAddr = c.vip.GetString("net.register-addr")
	c.RpcPort = c.vip.GetString("net.rpc-port")
	c.Target = c.vip.GetString("net.target") //optional
	c.Port = c.vip.GetInt64("net.port")
	c.MetricsPort = c.vip.GetString("net.metrics-port")
	c.AccSecKey = c.vip.GetString("account.sec-key")

	return nil
}
