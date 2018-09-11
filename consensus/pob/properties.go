package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
)

var staticProperty *StaticProperty

// StaticProperty handles the the static property of pob.
type StaticProperty struct {
	account           *account.Account
	NumberOfWitnesses int64
	WitnessList       []string
	Watermark         map[string]int64
}

func newStaticProperty(account *account.Account, witnessList []string) *StaticProperty {
	property := &StaticProperty{
		account:     account,
		WitnessList: make([]string, 0),
		Watermark:   make(map[string]int64),
	}

	property.updateWitness(witnessList)

	return property
}

func (property *StaticProperty) updateWitness(witnessList []string) {

	property.NumberOfWitnesses = int64(len(witnessList))
	property.WitnessList = witnessList
}

var (
	second2nanosecond int64 = 1000000000
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
	currentSlot := timeSec / (second2nanosecond * common.SlotLength)
	return (currentSlot+1)*second2nanosecond*common.SlotLength - timeSec
}
