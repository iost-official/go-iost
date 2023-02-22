package version

import "github.com/iost-official/go-iost/v3/common"

// ChainIDs
const (
	MainNetChainID uint32 = 1024
	TestNetChainID uint32 = 1023
)

// ChainConfig ...
type ChainConfig struct {
	Block3_9_0  int64
	Block3_10_0 int64
}

var (
	mainNetChainConf = &ChainConfig{
		Block3_9_0:  220000000,
		Block3_10_0: 520000000,
	}

	testNetChainConf = &ChainConfig{
		Block3_9_0:  0,
		Block3_10_0: 0,
	}

	defaultChainConf = &ChainConfig{
		Block3_9_0:  0,
		Block3_10_0: 0,
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

// IsFork3_9_0 ...
func IsFork3_9_0(num int64) bool {
	return isForked(chainConf.Block3_9_0, num)
}

// IsFork3_10_0 ...
func IsFork3_10_0(num int64) bool {
	return isForked(chainConf.Block3_10_0, num)
}

func isForked(v, num int64) bool {
	return v <= num
}

// Rules wraps original IsXxx functions
type Rules struct {
	IsFork3_9_0  bool `json:"is_fork3_9_0"`
	IsFork3_10_0 bool `json:"is_fork3_10_0"`
}

// NewRules create Rules for each block
func NewRules(num int64) *Rules {
	return &Rules{
		IsFork3_9_0:  IsFork3_9_0(num),
		IsFork3_10_0: IsFork3_10_0(num),
	}
}
