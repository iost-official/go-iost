package pob

import (
	. "github.com/iost-official/Go-IOS-Protocol/account"
	. "github.com/iost-official/Go-IOS-Protocol/consensus/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
)

var staticProp globalStaticProperty
var dynamicProp globalDynamicProperty

type globalStaticProperty struct {
	Account
	NumberOfWitnesses  int
	WitnessList        []string
	Watermark		   map[string]uint64
	SlotMap			   map[uint64]map[string]bool
	Producing		   bool
}

func newGlobalStaticProperty(acc Account, witnessList []string) globalStaticProperty {
	prop := globalStaticProperty{
		Account:            acc,
		NumberOfWitnesses:  len(witnessList),
		WitnessList:        witnessList,
		Watermark: 			make(map[string]uint64),
		SlotMap: 			make(map[uint64]map[string]bool),
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

const (
	slotPerWitness      = 1
	maintenanceInterval = 24
)

type globalDynamicProperty struct {
	LastBlockNumber          int64
	LastBlockTime            Timestamp
	LastBLockHash            []byte
	TotalSlots               int64
	LastConfirmedBlockNumber int32
	NextMaintenanceTime      Timestamp
}

func newGlobalDynamicProperty() globalDynamicProperty {
	prop := globalDynamicProperty{
		LastBlockNumber:          0,
		LastBlockTime:            Timestamp{Slot: 0},
		TotalSlots:               0,
		LastConfirmedBlockNumber: 0,
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
	prop.LastBlockTime = Timestamp{Slot: blockHead.Time}
	copy(prop.LastBLockHash, blockHead.Hash())
}

func (prop *globalDynamicProperty) timestampToSlot(time Timestamp) int64 {
	return time.Slot
}

func (prop *globalDynamicProperty) slotToTimestamp(slot int64) *Timestamp {
	return &Timestamp{Slot: slot}
}

func witnessOfSec(sec int64) string {
	time := GetTimestamp(sec)
	return witnessOfTime(time)
}

func witnessOfTime(time Timestamp) string {

	currentSlot := dynamicProp.timestampToSlot(time)
	slotsEveryTurn := int64(staticProp.NumberOfWitnesses * slotPerWitness)
	index := ((currentSlot % slotsEveryTurn) + slotsEveryTurn) % slotsEveryTurn
	index /= slotPerWitness
	witness := staticProp.WitnessList[index]

	return witness
}

func timeUntilNextSchedule(timeSec int64) int64 {
	var index int
	if index = getIndex(staticProp.Account.GetId(), staticProp.WitnessList); index < 0 {
		return dynamicProp.NextMaintenanceTime.ToUnixSec()
	}

	time := GetTimestamp(timeSec)
	currentSlot := dynamicProp.timestampToSlot(time)
	slotsEveryTurn := int64(staticProp.NumberOfWitnesses * slotPerWitness)
	k := currentSlot / slotsEveryTurn
	startSlot := k*slotsEveryTurn + int64(index*slotPerWitness)
	if startSlot > currentSlot {
		return dynamicProp.slotToTimestamp(startSlot).ToUnixSec() - timeSec
	}
	if currentSlot-startSlot < slotPerWitness {
		if time.Slot > dynamicProp.LastBlockTime.Slot {
			return 0
		} else if currentSlot+1 < startSlot+slotPerWitness {
			return dynamicProp.slotToTimestamp(currentSlot+1).ToUnixSec() - timeSec
		}
	}
	nextSlot := (k+1)*slotsEveryTurn + int64(index*slotPerWitness)
	return dynamicProp.slotToTimestamp(nextSlot).ToUnixSec() - timeSec
}
