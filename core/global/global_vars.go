package global

import (
	"github.com/iost-official/go-iost/common"
)

// BuildTime build time
var BuildTime string

// GitHash git hash
var GitHash string

var globalConf *common.Config

// SetGlobalConf ...
func SetGlobalConf(conf *common.Config) {
	globalConf = conf
}

// GetGlobalConf ...
func GetGlobalConf() *common.Config {
	return globalConf
}
