package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/common"
)

var staticProperty globalStaticProperty
var dynamicProperty globalDynamicProperty

type globalStaticProperty struct {
	account.Account
	NumberOfWitnesses int
	WitnessList       []string
	Watermark         map[string]uint64
	SlotMap           map[uint64]map[string]bool
}

func newGlobalStaticProperty(acc account.Account, witnessList []string) globalStaticProperty {
	prop := globalStaticProperty{
		Account:           acc,
		NumberOfWitnesses: len(witnessList),
		WitnessList:       witnessList,
		Watermark:         make(map[string]uint64),
		SlotMap:           make(map[uint64]map[string]bool),
	}
	return prop
}

func (prop *globalStaticProperty) updateWitnessList(newList []string) {
	prop.WitnessList = newList
	prop.NumberOfWitnesses = len(newList)
}

func (prop *globalStaticProperty) hasSlotWitness(slot uint64, witness string) bool {
	if prop.SlotMap[slot] == nil {
		return false
	} else {
		return prop.SlotMap[slot][witness]
	}
}

func (prop *globalStaticProperty) addSlotWitness(slot uint64, witness string) {
	if prop.SlotMap[slot] == nil {
		prop.SlotMap[slot] = make(map[string]bool)
	}
	prop.SlotMap[slot][witness] = true
}

func (prop *globalStaticProperty) delSlotWitness(slotStart uint64, slotEnd uint64) {
	for slot := slotStart; slot <= slotEnd; slot++ {
		if _, has := prop.SlotMap[slot]; has {
			delete(prop.SlotMap, slot)
		}
	}
}

func getIndex(element string, list []string) int {
	for index, ele := range list {
		if ele == element {
			return index
		}
	}
	return -1
}

var (
	slotPerWitness      = 1
	maintenanceInterval = 24
)

type globalDynamicProperty struct {
	LastBlockNumber     int64
	LastBlockTime       common.Timestamp
	LastBLockHash       []byte
	TotalSlots          int64
	NextMaintenanceTime common.Timestamp
}

func newGlobalDynamicProperty() globalDynamicProperty {
	prop := globalDynamicProperty{
		LastBlockNumber: 0,
		LastBlockTime:   common.Timestamp{Slot: 0},
		TotalSlots:      0,
	}
	prop.NextMaintenanceTime.AddHour(maintenanceInterval)
	return prop
}

func (prop *globalDynamicProperty) update(blockHead *block.BlockHead) {
	if prop.LastBlockNumber == 0 {
		prop.TotalSlots = 1
		prop.NextMaintenanceTime.AddHour(maintenanceInterval)
	}
	prop.LastBlockNumber = blockHead.Number
	prop.LastBlockTime = common.Timestamp{Slot: blockHead.Time}
	hash, err := blockHead.Hash()
	if err == nil {
		copy(prop.LastBLockHash, hash)
	}
}

func (prop *globalDynamicProperty) timestampToSlot(time common.Timestamp) int64 {
	return time.Slot
}

func (prop *globalDynamicProperty) slotToTimestamp(slot int64) *common.Timestamp {
	return &common.Timestamp{Slot: slot}
}

func witnessOfSec(sec int64) string {
	time := common.GetTimestamp(sec)
	return witnessOfTime(time)
}

func witnessOfTime(time common.Timestamp) string {

	currentSlot := dynamicProperty.timestampToSlot(time)
	slotsEveryTurn := int64(staticProperty.NumberOfWitnesses * slotPerWitness)
	index := currentSlot % slotsEveryTurn
	index /= int64(slotPerWitness)
	witness := staticProperty.WitnessList[index]

	return witness
}

func timeUntilNextSchedule(timeSec int64) int64 {
	var index int
	if index = getIndex(staticProperty.Account.GetId(), staticProperty.WitnessList); index < 0 {
		return dynamicProperty.NextMaintenanceTime.ToUnixSec()
	}

	time := common.GetTimestamp(timeSec)
	currentSlot := dynamicProperty.timestampToSlot(time)
	slotsEveryTurn := int64(staticProperty.NumberOfWitnesses * slotPerWitness)
	k := currentSlot / slotsEveryTurn
	startSlot := k*slotsEveryTurn + int64(index*slotPerWitness)
	if startSlot > currentSlot {
		return dynamicProperty.slotToTimestamp(startSlot).ToUnixSec() - timeSec
	}
	if currentSlot-startSlot < int64(slotPerWitness) {
		if time.Slot > dynamicProperty.LastBlockTime.Slot {
			return 0
		} else if currentSlot+1 < startSlot+int64(slotPerWitness) {
			return dynamicProperty.slotToTimestamp(currentSlot+1).ToUnixSec() - timeSec
		}
	}
	nextSlot := (k+1)*slotsEveryTurn + int64(index*slotPerWitness)
	return dynamicProperty.slotToTimestamp(nextSlot).ToUnixSec() - timeSec
}
