package pob2

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
	// 每个witness做几个slot以后换下一个
	slotPerWitness = 1
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
	prop:=globalDynamicProperty{
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
	//else {
	//	prop.TotalSlots = prop.timestampToSlot(Timestamp{blockHead.Time}) + 1
	//}
	prop.LastBlockNumber = blockHead.Number
	prop.LastBlockTime = Timestamp{Slot: blockHead.Time}
	copy(prop.LastBLockHash, blockHead.BlockHash)
}

func (prop *globalDynamicProperty) timestampToSlot(time Timestamp) int64 {
	return time.Slot
}

func (prop *globalDynamicProperty) slotToTimestamp(slot int64) *Timestamp {
	return &Timestamp{Slot: slot}
}

// 返回对于指定的Unix时间点，应该轮到生产块的节点id
func witnessOfSec(sp *globalStaticProperty, dp *globalDynamicProperty, sec int64) string {
	time := GetTimestamp(sec)
	return witnessOfTime(sp, dp, time)
}

// 返回对于指定的时间戳，应该轮到生产块的节点id
func witnessOfTime(sp *globalStaticProperty, dp *globalDynamicProperty, time Timestamp) string {

	currentSlot := dp.timestampToSlot(time)
	slotsEveryTurn := int64(sp.NumberOfWitnesses * slotPerWitness)
	index := ((currentSlot % slotsEveryTurn) + slotsEveryTurn) % slotsEveryTurn
	index /= slotPerWitness
	witness := sp.WitnessList[index]

	return witness
}

// 返回到下一次轮到本节点生产块的时间长度，秒为单位
// 如果该节点当前不是witness，则返回到达下一次maintenance的时间长度
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
	// 当前还没到本轮本节点的起始slot
	if startSlot > currentSlot {
		return dp.slotToTimestamp(startSlot).ToUnixSec() - timeSec
	}
	// 当前slot在本轮中
	if currentSlot-startSlot < slotPerWitness {
		if time.Slot > dp.LastBlockTime.Slot {
			// 当前slot还未产生块，需要立即产生块
			// TODO: 考虑线程间同步问题，使用另外的方法判断当前slot是否产生块
			return 0
		} else if currentSlot+1 < startSlot+slotPerWitness {
			// 当前slot已经产生块，并且下一个slot还是本节点产生
			return dp.slotToTimestamp(currentSlot+1).ToUnixSec() - timeSec
		}
	}
	// 本轮本节点已经产生完毕，需要等到下一轮产生
	nextSlot := (k+1)*slotsEveryTurn + int64(index*slotPerWitness)
	return dp.slotToTimestamp(nextSlot).ToUnixSec() - timeSec
}
