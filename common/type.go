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
	SlotInterval       = 3 * time.Second
	BlockInterval      = 500 * time.Millisecond
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
	slot := nanosec / int64(SlotInterval)
	index := slot % int64(len(witnessList))
	witness := witnessList[index]
	return witness
}

// SlotOfUnixNano will return the slot number of unixnano.
func SlotOfUnixNano(unixnano int64) int64 {
	return unixnano / int64(SlotInterval)
}

// NextSlotTime will return the time in the next slot.
func NextSlotTime() time.Time {
	currentSlot := time.Now().UnixNano() / int64(SlotInterval)
	nextSlotUnixNano := (currentSlot + 1) * int64(SlotInterval)
	nextSlotTime := time.Unix(nextSlotUnixNano/int64(time.Second), nextSlotUnixNano%int64(time.Second))
	ilog.Debugf("The next slot: %v", nextSlotTime.UnixNano())
	return nextSlotTime
}

// NextSlot will return the slot number in the next slot.
func NextSlot() int64 {
	return time.Now().UnixNano()/int64(SlotInterval) + 1
}

func TimeOfBlock(slot int64, num int64) time.Time {
	unixNano := slot*int64(SlotInterval) + num*int64(BlockInterval)
	return time.Unix(unixNano/int64(time.Second), unixNano%int64(time.Second))
}
