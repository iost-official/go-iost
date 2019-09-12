package version

import "github.com/iost-official/go-iost/common"

// ChainIDs
const (
	MainNetChainID uint32 = 1024
	TestNetChainID uint32 = 1023
)

// ChainConfig ...
type ChainConfig struct {
	Block3_0_10 int64
	Block3_1_0  int64
	Block3_3_0  int64
}

var (
	mainNetChainConf = &ChainConfig{
		Block3_0_10: 12000000,
		Block3_1_0:  15800000,
		Block3_3_0:  18000000,
	}

	testNetChainConf = &ChainConfig{
		Block3_0_10: 10599000,
		Block3_1_0:  12800000,
		Block3_3_0:  30440000,
	}

	defaultChainConf = &ChainConfig{
		Block3_0_10: 0,
		Block3_1_0:  0,
		Block3_3_0:  0,
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

// IsFork3_3_0 ...
func IsFork3_3_0(num int64) bool {
	return isForked(chainConf.Block3_3_0, num)
}

// IsFork3_1_0 ...
func IsFork3_1_0(num int64) bool {
	return isForked(chainConf.Block3_1_0, num)
}

// IsFork3_0_10 ...
func IsFork3_0_10(num int64) bool {
	return isForked(chainConf.Block3_0_10, num)
}

func isForked(v, num int64) bool {
	return v <= num
}

// Rules wraps original IsXxx functions
type Rules struct {
	IsFork3_0_10 bool `json:"is_fork3_0_10"`
	IsFork3_1_0  bool `json:"is_fork3_1_0"`
	IsFork3_3_0  bool `json:"is_fork3_3_0"`
}

// NewRules create Rules for each block
func NewRules(num int64) *Rules {
	return &Rules{
		IsFork3_0_10: IsFork3_0_10(num),
		IsFork3_1_0:  IsFork3_1_0(num),
		IsFork3_3_0:  IsFork3_3_0(num),
	}
}
