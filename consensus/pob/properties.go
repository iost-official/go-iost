package pob

import (
	"errors"
	"sort"

	. "github.com/iost-official/prototype/account"
	. "github.com/iost-official/prototype/consensus/common"
	"github.com/iost-official/prototype/core/block"
)

type globalStaticProperty struct {
	Account
	NumberOfWitnesses  int
	WitnessList        []string
	PendingWitnessList []string
}

func newGlobalStaticProperty(acc Account, witnessList []string) globalStaticProperty {
	prop := globalStaticProperty{
		Account:            acc,
		NumberOfWitnesses:  len(witnessList),
		WitnessList:        witnessList,
		PendingWitnessList: []string{},
	}
	return prop
}

func (prop *globalStaticProperty) addPendingWitness(id string) error {
	for _, wit := range prop.WitnessList {
		if id == wit {
			return errors.New("already in witness list")
		}
	}
	for _, wit := range prop.PendingWitnessList {
		if id == wit {
			return errors.New("already in pending list")
		}
	}
	prop.PendingWitnessList = append(prop.PendingWitnessList, id)
	return nil
}

func (prop *globalStaticProperty) deletePendingWitness(id string) error {
	i := 0
	for _, wit := range prop.PendingWitnessList {
		if id == wit {
			newList := append(prop.PendingWitnessList[:i], prop.PendingWitnessList[i+1:]...)
			prop.PendingWitnessList = newList
			return nil
		}
		i++
	}
	return errors.New("witness not in pending list")
}

func (prop *globalStaticProperty) updateWitnessLists(newList []string) {
	var newPendingList []string
	for _, wit := range prop.WitnessList {
		if !inList(wit, newList) {
			newPendingList = append(newPendingList, wit)
		}
	}
	for _, wit := range prop.PendingWitnessList {
		if !inList(wit, newList) {
			newPendingList = append(newPendingList, wit)
		}
	}
	sort.Strings(newList)
	prop.WitnessList = newList
	prop.PendingWitnessList = newPendingList
}

func inList(element string, list []string) bool {
	for _, ele := range list {
		if ele == element {
			return true
		}
	}
	return false
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

func witnessOfSec(sp *globalStaticProperty, dp *globalDynamicProperty, sec int64) string {
	time := GetTimestamp(sec)
	return witnessOfTime(sp, dp, time)
}

func witnessOfTime(sp *globalStaticProperty, dp *globalDynamicProperty, time Timestamp) string {

	currentSlot := dp.timestampToSlot(time)
	slotsEveryTurn := int64(sp.NumberOfWitnesses * slotPerWitness)
	index := ((currentSlot % slotsEveryTurn) + slotsEveryTurn) % slotsEveryTurn
	index /= slotPerWitness
	witness := sp.WitnessList[index]

	return witness
}

func timeUntilNextSchedule(sp *globalStaticProperty, dp *globalDynamicProperty, timeSec int64) int64 {
	var index int
	if index = getIndex(sp.Account.GetId(), sp.WitnessList); index < 0 {
		return dp.NextMaintenanceTime.ToUnixSec()
	}

	time := GetTimestamp(timeSec)
	currentSlot := dp.timestampToSlot(time)
	slotsEveryTurn := int64(sp.NumberOfWitnesses * slotPerWitness)
	k := currentSlot / slotsEveryTurn
	startSlot := k*slotsEveryTurn + int64(index*slotPerWitness)
	if startSlot > currentSlot {
		return dp.slotToTimestamp(startSlot).ToUnixSec() - timeSec
	}
	if currentSlot-startSlot < slotPerWitness {
		if time.Slot > dp.LastBlockTime.Slot {
			return 0
		} else if currentSlot+1 < startSlot+slotPerWitness {
			return dp.slotToTimestamp(currentSlot+1).ToUnixSec() - timeSec
		}
	}
	nextSlot := (k+1)*slotsEveryTurn + int64(index*slotPerWitness)
	return dp.slotToTimestamp(nextSlot).ToUnixSec() - timeSec
}
