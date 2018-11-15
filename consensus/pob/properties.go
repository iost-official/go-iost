package pob

import (
	"strings"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
)

var staticProperty *StaticProperty

// StaticProperty handles the the static property of pob.
type StaticProperty struct {
	account           *account.KeyPair
	NumberOfWitnesses int64
	WitnessList       []string
	Watermark         map[string]int64
}

func newStaticProperty(account *account.KeyPair, witnessList []string) *StaticProperty {
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

func (property *StaticProperty) isWitness(w string) bool {
	for _, v := range property.WitnessList {
		if strings.Compare(v, w) == 0 {
			return true
		}
	}
	return false
}

var (
	second2nanosecond int64 = 1000000000
)

func witnessOfNanoSec(nanosec int64) string {
	return witnessOfSec(nanosec / second2nanosecond)
}

func witnessOfSec(sec int64) string {
	return witnessOfSlot(sec / common.SlotLength)
}

func witnessOfSlot(slot int64) string {
	index := slot % staticProperty.NumberOfWitnesses
	witness := staticProperty.WitnessList[index]
	ilog.Info(index, staticProperty.WitnessList, witness)
	return witness
}

func timeUntilNextSchedule(timeSec int64) int64 {
	currentSlot := timeSec / (second2nanosecond * common.SlotLength)
	return (currentSlot+1)*second2nanosecond*common.SlotLength - timeSec
}

// GetStaticProperty return property. RPC needs it.
func GetStaticProperty() *StaticProperty {
	return staticProperty
}
