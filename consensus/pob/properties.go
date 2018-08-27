package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
)

var staticProperty *StaticProperty

// StaticProperty handles the the static property of pob.
type StaticProperty struct {
	account           account.Account
	NumberOfWitnesses int64
	WitnessList       []string
	WitnessMap        map[string]int64
	Watermark         map[string]int64
	SlotMap           map[int64]bool
}

func newStaticProperty(account account.Account, witnessList []string) *StaticProperty {
	property := &StaticProperty{
		account:           account,
		NumberOfWitnesses: int64(len(witnessList)),
		WitnessList:       witnessList,
		WitnessMap:        make(map[string]int64),
		Watermark:         make(map[string]int64),
		SlotMap:           make(map[int64]bool),
	}
	for i, w := range witnessList {
		property.WitnessMap[w] = int64(i)
	}
	return property
}

func (property *StaticProperty) hasSlot(slot int64) bool {
	return property.SlotMap[slot]
}

func (property *StaticProperty) addSlot(slot int64) {
	property.SlotMap[slot] = true
}

func (property *StaticProperty) delSlot(slot int64) {

	if slot%10 != 0 {
		return
	}

	for k := range property.SlotMap {
		if k <= slot {
			delete(property.SlotMap, k)
		}
	}
}

var (
	second2nanosecond   int64 = 1000000000
	maintenanceInterval       = 24 * second2nanosecond
)

func witnessOfSec(sec int64) string {
	return witnessOfSlot(sec / common.SlotLength)
}

func witnessOfSlot(slot int64) string {
	index := slot % staticProperty.NumberOfWitnesses
	witness := staticProperty.WitnessList[index]
	return witness
}

func timeUntilNextSchedule(timeSec int64) int64 {
	index, ok := staticProperty.WitnessMap[staticProperty.account.ID]
	if !ok {
		return maintenanceInterval * common.SlotLength
	}
	currentSlot := timeSec / (second2nanosecond * common.SlotLength)
	round := currentSlot / staticProperty.NumberOfWitnesses
	startSlot := round*staticProperty.NumberOfWitnesses + index
	nextSlot := (round+1)*staticProperty.NumberOfWitnesses + index
	if currentSlot > startSlot {
		return nextSlot*common.SlotLength*second2nanosecond - timeSec
	} else if currentSlot < startSlot {
		return startSlot*common.SlotLength*second2nanosecond - timeSec
	} else {
		return 0
	}
}
