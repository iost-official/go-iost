package pob

import (
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
	"strings"
	"sync"
)

var staticProperty *StaticProperty

// StaticProperty handles the the static property of pob.
type StaticProperty struct {
	account           *account.KeyPair
	NumberOfWitnesses int64
	witnessList       []string
	Watermark         map[string]int64
	SlotUsed          map[int64]bool
	mu                sync.RWMutex
}

func newStaticProperty(account *account.KeyPair, witnessList []string) *StaticProperty {
	property := &StaticProperty{
		account:     account,
		witnessList: make([]string, 0),
		Watermark:   make(map[string]int64),
		SlotUsed:    make(map[int64]bool),
	}

	property.updateWitness(witnessList)

	return property
}

// WitnessList is return witnessList
func (p *StaticProperty) WitnessList() []string {
	return p.witnessList
}

func (p *StaticProperty) updateWitness(witnessList []string) {
	defer p.mu.Unlock()
	p.mu.Lock()

	p.NumberOfWitnesses = int64(len(witnessList))
	p.witnessList = witnessList
}

func (p *StaticProperty) witnessByIndex(i int) string {
	defer p.mu.RUnlock()
	p.mu.RLock()

	if i >= len(p.witnessList) {
		ilog.Errorf("witnessByIndex index %v is error", i)
		return ""
	}

	return p.witnessList[i]
}

func (p *StaticProperty) isWitness(w string) bool {
	defer p.mu.RUnlock()
	p.mu.RLock()

	for _, v := range p.witnessList {
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
	witness := staticProperty.witnessByIndex(int(index))
	return witness
}

func slotOfSec(sec int64) int64 {
	return sec / common.SlotLength
}

//func timeUntilNextSchedule(timeSec int64) int64 {
//	currentSlot := timeSec / (second2nanosecond * common.SlotLength)
//	return (currentSlot+1)*second2nanosecond*common.SlotLength - timeSec
//}

// GetStaticProperty return property. RPC needs it.
func GetStaticProperty() *StaticProperty {
	return staticProperty
}
