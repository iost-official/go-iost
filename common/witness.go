
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

// NextSlot will return the slot number in the next slot.
func NextSlot() int64 {
	return time.Now().UnixNano()/int64(SlotInterval) + 1
}

// TimeOfBlock will return the block time for specific slots and num.
func TimeOfBlock(slot int64, num int64) time.Time {
	unixNano := slot*int64(SlotInterval) + num*int64(BlockInterval)
	return time.Unix(unixNano/int64(time.Second), unixNano%int64(time.Second))
}

