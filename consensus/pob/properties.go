package pob

import (
	"github.com/iost-official/go-iost/common"
)

var (
	second2nanosecond int64 = 1000000000
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
	return witnessOfSec(nanosec/second2nanosecond, witnessList)
}

func witnessOfSec(sec int64, witnessList []string) string {
	return witnessOfSlot(sec/common.SlotLength, witnessList)
}

func witnessOfSlot(slot int64, witnessList []string) string {
	index := slot % int64(len(witnessList))
	witness := witnessList[index]
	return witness
}

func slotOfSec(sec int64) int64 {
	return sec / common.SlotLength
}

func timeUntilNextSchedule(timeSec int64) int64 {
	currentSlot := timeSec / (second2nanosecond * common.SlotLength)
	return (currentSlot+1)*second2nanosecond*common.SlotLength - timeSec
}
