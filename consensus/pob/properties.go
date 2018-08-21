package pob

import (
	"fmt"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
)

var staticProperty StaticProperty

type StaticProperty struct {
	account           account.Account
	NumberOfWitnesses int64
	WitnessList       []string
	WitnessMap        map[string]int64
	Watermark         map[string]int64
	SlotMap           map[int64]bool
}

func newStaticProperty(account account.Account, witnessList []string) StaticProperty {
	property := StaticProperty{
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

func (property *StaticProperty) updateWitnessList(witnessList []string) {
	property.WitnessList = witnessList
	for i, w := range witnessList {
		property.WitnessMap[w] = int64(i)
	}
	property.NumberOfWitnesses = int64(len(witnessList))
}

func (property *StaticProperty) hasSlot(slot int64) bool {
	return property.SlotMap[slot]
}

func (property *StaticProperty) addSlot(slot int64) {
	property.SlotMap[slot] = true
}

var (
	maintenanceInterval int64 = 24
)

func witnessOfSec(sec int64) string {
	return witnessOfSlot(sec / common.SlotLength)
}

func witnessOfSlot(slot int64) string {
	index := slot % staticProperty.NumberOfWitnesses
	fmt.Println(index)
	witness := staticProperty.WitnessList[index]
	return witness
}

func timeUntilNextSchedule(timeSec int64) int64 {
	index, ok := staticProperty.WitnessMap[staticProperty.account.ID]
	if !ok {
		return maintenanceInterval * common.SlotLength
	}
	currentSlot := timeSec / common.SlotLength
	round := currentSlot / staticProperty.NumberOfWitnesses
	startSlot := round*staticProperty.NumberOfWitnesses + index
	if currentSlot > startSlot {
		nextSlot := (round+1)*staticProperty.NumberOfWitnesses + index
		return nextSlot*common.SlotLength - timeSec
	} else if currentSlot < startSlot {
		return startSlot*common.SlotLength - timeSec
	} else {
		return 0
	}
}
