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

func NewGlobalStaticProperty(member core.Member, witnessList []string) *GlobalStaticProperty {
	prop := &GlobalStaticProperty{
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
