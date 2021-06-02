package common

import (
	"sync"

	"github.com/iost-official/go-iost/v3/metrics"
)

// ModeType is the type of mode.
type ModeType uint

// Constant of mode type.
const (
	ModeNormal ModeType = iota
	ModeSync
	ModeInit
)

var (
	mode      = ModeInit
	modeMutex = new(sync.RWMutex)
	modeGauge = metrics.NewGauge("iost_node_mode", nil)
)

func init() {
	modeGauge.Set(float64(mode), nil)
}

// Mode will return the mode of iserver.
func Mode() string {
	modeMutex.RLock()
	defer modeMutex.RUnlock()

	switch mode {
	case ModeNormal:
		return "ModeNormal"
	case ModeSync:
		return "ModeSync"
	case ModeInit:
		return "ModeInit"
	default:
		return "Undefined"
	}
}

// SetMode will set the mode of iserver.
func SetMode(m ModeType) {
	modeMutex.Lock()
	defer modeMutex.Unlock()

	mode = m
	modeGauge.Set(float64(mode), nil)
}
