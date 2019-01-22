package pob

import (
	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
)

var staticProperty *StaticProperty

// StaticProperty handles the the static property of pob.
type StaticProperty struct {
	account           *account.KeyPair
	NumberOfWitnesses int64
	SlotUsed          map[int64]bool
}

func newStaticProperty(account *account.KeyPair, number int64) *StaticProperty {
	property := &StaticProperty{
		account:           account,
		NumberOfWitnesses: number,
		SlotUsed:          make(map[int64]bool),
	}
	return property
}

func (property *StaticProperty) isWitness(w string, witnessList []string) bool {
	for _, v := range witnessList {
		if strings.Compare(v, w) == 0 {
			return true
		}
	}
	return false
}

var (
	second2nanosecond int64 = 1000000000
)

func witnessOfNanoSec(nanosec int64, witnessList []string) string {
	return witnessOfSec(nanosec/second2nanosecond, witnessList)
}

func witnessOfSec(sec int64, witnessList []string) string {
	return witnessOfSlot(sec/common.SlotLength, witnessList)
}

func witnessOfSlot(slot int64, witnessList []string) string {
	index := slot % staticProperty.NumberOfWitnesses
	witness := witnessList[index]
	return witness
}

func timeUntilNextSchedule(timeSec int64) int64 {
	currentSlot := timeSec / (second2nanosecond * common.SlotLength)
	return (currentSlot+1)*second2nanosecond*common.SlotLength - timeSec
}

// GetStaticProperty return property. RPC needs it.
func GetStaticProperty() *StaticProperty {
	return staticProperty
}
