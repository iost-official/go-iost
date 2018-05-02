package dpos

import (
	"errors"
	"sort"

	"github.com/iost-official/prototype/core"
)

type GlobalStaticProperty struct {
	Member             core.Member
	NumberOfWitnesses  int
	WitnessList        []string
	PendingWitnessList []string
}

func NewGlobalStaticProperty(member core.Member, witnessList []string) GlobalStaticProperty {
	prop := GlobalStaticProperty{
		Member:             member,
		NumberOfWitnesses:  len(witnessList),
		WitnessList:        witnessList,
		PendingWitnessList: []string{},
	}
	return prop
}

func (prop *GlobalStaticProperty) AddPendingWitness(id string) error {
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

func (prop *GlobalStaticProperty) DeletePendingWitness(id string) error {
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

func (prop *GlobalStaticProperty) UpdateWitnessLists(newList []string) {
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
	SlotPerWitness = 1
)

type GlobalDynamicProperty struct {
	LastBlockNumber          int32
	LastBlockTime            core.Timestamp
	LastBLockHash            []byte
	TotalSlots               int64
	LastConfirmedBlockNumber int32
	NextMaintenanceTime      core.Timestamp
}

func NewGlobalDynamicProperty() GlobalDynamicProperty {
	return GlobalDynamicProperty{
		LastBlockNumber:          0,
		LastBlockTime:            core.Timestamp{0},
		TotalSlots:               0,
		LastConfirmedBlockNumber: 0,
		NextMaintenanceTime:      core.Timestamp{0},
	}
}

func (prop *GlobalDynamicProperty) Update(blockHead *core.BlockHead) {
	if prop.LastBlockNumber == 0 {
		prop.TotalSlots = 1
		prop.NextMaintenanceTime.AddHour(MaintenanceInterval)
	} else {
		prop.TotalSlots = prop.timestampToSlot(blockHead.Time) + 1
	}
	prop.LastBlockNumber = blockHead.Number
	prop.LastBlockTime = blockHead.Time
	copy(prop.LastBLockHash, blockHead.BlockHash)
}

func (prop *GlobalDynamicProperty) timestampToSlot(time core.Timestamp) int64 {
	return time.Slot - prop.LastBlockTime.Slot + prop.TotalSlots - 1
}

func (prop *GlobalDynamicProperty) slotToTimestamp(slot int64) *core.Timestamp {
	return &core.Timestamp{Slot: slot - prop.TotalSlots + prop.LastBlockTime.Slot + 1}
}

// 返回对于指定的Unix时间点，应该轮到生产块的节点id
func WitnessOfSec(sp *GlobalStaticProperty, dp *GlobalDynamicProperty, sec int64) string {
	time := core.GetTimestamp(sec)
	return WitnessOfTime(sp, dp, time)
}

// 返回对于指定的时间戳，应该轮到生产块的节点id
func WitnessOfTime(sp *GlobalStaticProperty, dp *GlobalDynamicProperty, time core.Timestamp) string {
	// 当前一个块是创世块，应该让第一个witness产生块
	// 问题：如果第一个witness失败？
	if dp.LastBlockNumber == 0 {
		return sp.WitnessList[0]
	}
	currentSlot := dp.timestampToSlot(time)
	index := currentSlot % int64(sp.NumberOfWitnesses*SlotPerWitness)
	index /= SlotPerWitness
	return sp.WitnessList[index]
}

// 返回到下一次轮到本节点生产块的时间长度，秒为单位
// 如果该节点当前不是witness，则返回到达下一次maintenance的时间长度
func TimeUntilNextSchedule(sp *GlobalStaticProperty, dp *GlobalDynamicProperty, timeSec int64) int64 {
	var index int
	if index = getIndex(sp.Member.GetId(), sp.WitnessList); index < 0 {
		return dp.NextMaintenanceTime.ToUnixSec() - timeSec
	}
	// 如果上一个块是创世块，并且本节点应该首先产生块，无需计算时间
	if dp.LastBlockNumber == 0 && index == 0 {
		return 0
	}
	time := core.GetTimestamp(timeSec)
	currentSlot := dp.timestampToSlot(time)
	slotsEveryTurn := int64(sp.NumberOfWitnesses * SlotPerWitness)
	k := currentSlot / slotsEveryTurn
	startSlot := k*slotsEveryTurn + int64(index*SlotPerWitness)
	// 当前还没到本轮本节点的起始slot
	if startSlot > currentSlot {
		return dp.slotToTimestamp(startSlot).ToUnixSec() - timeSec
	}
	// 当前slot在本轮中
	if currentSlot-startSlot < SlotPerWitness {
		if time.Slot > dp.LastBlockTime.Slot {
			// 当前slot还未产生块，需要立即产生块
			// TODO: 考虑线程间同步问题，使用另外的方法判断当前slot是否产生块
			return 0
		} else if currentSlot+1 < startSlot+SlotPerWitness {
			// 当前slot已经产生块，并且下一个slot还是本节点产生
			return dp.slotToTimestamp(currentSlot+1).ToUnixSec() - timeSec
		}
	}
	// 本轮本节点已经产生完毕，需要等到下一轮产生
	nextSlot := (k+1)*slotsEveryTurn + int64(index*SlotPerWitness)
	return dp.slotToTimestamp(nextSlot).ToUnixSec() - timeSec
}
