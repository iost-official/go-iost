package common

import (
	"sync"
	"time"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
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

// Witness
var (
	VoteInterval       = int64(1200)
	SlotTime           = 3 * time.Second
	BlockNumPerWitness = 6
)

// IsWitness will judage if a public key is a witness.
func IsWitness(w string, witnessList []string) bool {
	for _, v := range witnessList {
		if v == w {
			return true
		}
	}
	return false
}

// WitnessOfNanoSec will return which witness is the current time.
func WitnessOfNanoSec(nanosec int64, witnessList []string) string {
	slot := nanosec / int64(SlotTime)
	index := slot % int64(len(witnessList))
	witness := witnessList[index]
	return witness
}

// SlotOfNanoSec will return current slot number.
func SlotOfNanoSec(nanosec int64) int64 {
	return nanosec / int64(SlotTime)
}

// TimeUntilNextSchedule will return the time left in the next slot.
func TimeUntilNextSchedule() time.Duration {
	now := time.Duration(time.Now().UnixNano())
	currentSlot := now / SlotTime
	nextSchedule := (currentSlot+1)*SlotTime - now
	ilog.Debugf("The nextSchedule: %.2f", nextSchedule.Seconds())
	return nextSchedule
}

// NextSlotTime will return the time in the next slot.
func NextSlotTime() time.Time {
	currentSlot := time.Now().UnixNano() / int64(SlotTime)
	nextSlotUnixNano := (currentSlot + 1) * int64(SlotTime)
	nextSlotTime := time.Unix(nextSlotUnixNano/int64(time.Second), nextSlotUnixNano%int64(time.Second))
	ilog.Debugf("The next slot: %v", nextSlotTime)
	return nextSlotTime
}
