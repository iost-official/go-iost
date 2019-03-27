package pob

import (
	"github.com/iost-official/go-iost/common"
)

func isWitness(w string, witnessList []string) bool {
	for _, v := range witnessList {
		if v == w {
			return true
		}
	}
	return false
}

func witnessOfNanoSec(nanosec int64, witnessList []string) string {
	return witnessOfSlot(nanosec/int64(common.SlotTime), witnessList)
}

func witnessOfSlot(slot int64, witnessList []string) string {
	index := slot % int64(len(witnessList))
	witness := witnessList[index]
	return witness
}

func slotOfNanoSec(nanosec int64) int64 {
	return nanosec / int64(common.SlotTime)
}

func timeUntilNextSchedule(timeSec int64) int64 {
	currentSlot := timeSec / int64(common.SlotTime)
	return (currentSlot+1)*int64(common.SlotTime) - timeSec
}
