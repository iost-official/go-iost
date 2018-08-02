package common

import "github.com/spf13/viper"

type Config struct {
	ldbPath       string
	redisAddr     string
	redisPort     int64
	netLogPath    string
	nodeTablePath string
	nodeID        string
	listenAddr    string
	regAddr       string
	rpcPort       string
	target        string
	port          int64
	metricsPort   string
	accSecKey     string
	logPath       string
	cfgFile       string
	logFile       string
	dbFile        string
}

func NewConfig() (*Config, error) {

	return &Config{}, nil
}

func (c *Config) LocalConfig(vip *viper.Viper) error {

	c.cfgFile = vip.GetString("config")
	c.logFile = vip.GetString("log")
	c.dbFile = vip.GetString("db")

	c.logPath = vip.GetString("log.path")
	c.ldbPath = vip.GetString("ldb.path")
	c.redisAddr = vip.GetString("redis.addr")
	c.redisPort = vip.GetInt64("redis.port")
	c.netLogPath = vip.GetString("net.log-path")
	c.nodeTablePath = vip.GetString("net.node-table-path")
	c.nodeID = vip.GetString("net.node-id") //optional
	c.listenAddr = vip.GetString("net.listen-addr")
	c.regAddr = vip.GetString("net.register-addr")
	c.rpcPort = vip.GetString("net.rpc-port")
	c.target = vip.GetString("net.target") //optional
	c.port = vip.GetInt64("net.port")
	c.metricsPort = vip.GetString("net.metrics-port")
	c.accSecKey = vip.GetString("account.sec-key")

	return nil
}
