package common

import "time"

// consts
const (
	VoteInterval       = 1200
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
func TimeUntilNextSchedule(timeSec int64) int64 {
	currentSlot := timeSec / int64(SlotTime)
	return (currentSlot+1)*int64(SlotTime) - timeSec
}
