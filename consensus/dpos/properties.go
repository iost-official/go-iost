package dpos

import (
	"github.com/iost-official/prototype/core"
	"sort"
)

type GlobalStaticProperty struct {
	Id                 string
	NumberOfWitnesses  int
	WitnessList        []string
	PendingWitnessList []string
}

func NewGlobalStaticProperty(id string, witnessList []string) *GlobalStaticProperty {
	prop := &GlobalStaticProperty{
		Id:					id,
		NumberOfWitnesses:	len(witnessList),
		WitnessList:		witnessList,
		PendingWitnessList: []string{},
	}
	return prop
}

func (prop *GlobalStaticProperty) AddPendingWitness(id string) {
	for _, wit := range prop.PendingWitnessList {
		if id == wit {
			return
		}
	}
	prop.PendingWitnessList = append(prop.PendingWitnessList, id)
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

type GlobalDynamicProperty struct {
	LastBlockNumber          int32
	LastBlockTime            core.Timestamp
	LastBLockHash            []byte
	TotalSlots               int64
	LastConfirmedBlockNumber int32
	NextMaintenanceTime      core.Timestamp
}

func NewGlobalDynamicProperty() *GlobalDynamicProperty {
	return &GlobalDynamicProperty{}
}
