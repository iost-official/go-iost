package version

import "github.com/iost-official/go-iost/common"

// ChainIDs
const (
	MainNetChainID uint32 = 1024
	TestNetChainID uint32 = 1023
)

// ChainConfig ...
type ChainConfig struct {
	Block3_1_0 int64
}

var (
	mainNetChainConf = &ChainConfig{
		Block3_1_0: 14400000,
	}

	testNetChainConf = &ChainConfig{
		Block3_1_0: 10600000,
	}

	defaultChainConf = &ChainConfig{
		Block3_1_0: 0,
	}
)

var chainConf = defaultChainConf

// InitChainConf ...
func InitChainConf(conf *common.Config) {
	switch conf.P2P.ChainID {
	case MainNetChainID:
		chainConf = mainNetChainConf
	case TestNetChainID:
		chainConf = testNetChainConf
	default:
		chainConf = defaultChainConf
	}
}

// IsFork3_1_0 ...
func IsFork3_1_0(num int64) bool {
	return isForked(chainConf.Block3_1_0, num)
}

func isForked(v, num int64) bool {
	return v <= num
}

// Rules wraps original IsXxx functions
type Rules struct {
	IsFork3_1_0 bool
}

// NewRules create Rules for each block
func NewRules(num int64) *Rules {
	return &Rules{
		IsFork3_1_0: IsFork3_1_0(num),
	}
}
